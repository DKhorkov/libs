package logging

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// directoryPermission — права каталога логов, если его создаёт DailyFile.
	directoryPermission = 0o755

	// retryInterval — пауза перед новой попыткой открыть файл после сбоя.
	retryInterval = time.Minute
)

// errNoLogFile — файла нет: открыть не удалось, повтор ещё не наступил.
var errNoLogFile = errors.New("log file is not open")

// DailyFile — io.Writer, пишущий в файл текущей даты: путь — pattern, где %s
// заменён датой в формате dateFormat и часовом поясе location (nil — UTC).
//
// В полночь файл закрывается и открывается новый: файл прошлого дня
// перестаёт расти, и удаление по возрасту (mtime) его находит. Путь,
// посчитанный один раз на старте, давал файл на весь запуск: он рос до
// перезапуска, и срок хранения его не трогал — в него писали. Перезапуск в тот
// же день дописывает файл дня. Каталог создаётся, если его нет. Дата
// считается только при переходе через полночь, а не на каждой строке.
//
// Файл не открылся — одно предупреждение в stdout на сбой и повтор не чаще
// раза в retryInterval; был открыт файл прошлой даты — строки идут в него.
type DailyFile struct {
	// mu — запись и смена файла из многих горутин логгера.
	mu sync.Mutex

	// pattern — путь к файлу с одним %s под дату.
	pattern string

	// dateFormat — формат даты в имени файла (раскладка time).
	dateFormat string

	// location — часовой пояс, в котором наступает полночь.
	location *time.Location

	// now — текущее время; подменяется в тестах.
	now func() time.Time

	// openFile — создание каталога и открытие файла; подменяется в тестах.
	openFile func(path string) (*os.File, error)

	// warnings — куда пишется предупреждение о сбое открытия.
	warnings io.Writer

	// date — дата открытого файла.
	date string

	// file — открытый файл даты date; nil до первой записи и после сбоя.
	file *os.File

	// switchAt — ближайшая полночь: до неё дата не пересчитывается.
	switchAt time.Time

	// retryAt — после сбоя открытия: раньше этого момента не пробовать снова.
	retryAt time.Time

	// warned — о текущем сбое уже предупредили.
	warned bool
}

// NewDailyFile — писатель в файл текущей даты. Файл открывается при первой
// записи.
func NewDailyFile(pattern, dateFormat string, location *time.Location) *DailyFile {
	if location == nil {
		location = time.UTC
	}

	return &DailyFile{ //nolint:exhaustruct // состояние файла заполняет первая запись
		pattern:    pattern,
		dateFormat: dateFormat,
		location:   location,
		now:        time.Now,
		openFile:   openLogFile,
		warnings:   os.Stdout,
	}
}

// Write пишет в файл текущей даты, в полночь переключаясь на новый.
func (d *DailyFile) Write(data []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if now := d.now(); d.file == nil || !now.Before(d.switchAt) {
		d.rotate(now)
	}

	if d.file == nil {
		return 0, errNoLogFile
	}

	return d.file.Write(data)
}

// Close закрывает открытый файл.
func (d *DailyFile) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.file == nil {
		return nil
	}

	err := d.file.Close()
	d.file = nil

	return err
}

// rotate открывает файл даты now вместо текущего, если пора. Прежний файл
// закрывается только после того, как новый открылся: при сбое он остаётся.
func (d *DailyFile) rotate(now time.Time) {
	if now.Before(d.retryAt) {
		return
	}

	local := now.In(d.location)
	date := local.Format(d.dateFormat)
	year, month, day := local.Date()
	switchAt := time.Date(year, month, day+1, 0, 0, 0, 0, d.location)

	if d.file != nil && date == d.date {
		d.switchAt = switchAt

		return
	}

	path := fmt.Sprintf(d.pattern, date)

	file, err := d.openFile(path)
	if err != nil {
		if !d.warned {
			_, _ = fmt.Fprintf(d.warnings, "Failed to open log file %s: %s\n", path, err)
			d.warned = true
		}

		d.retryAt = now.Add(retryInterval)

		return
	}

	if d.file != nil {
		_ = d.file.Close()
	}

	d.file = file
	d.date = date
	d.switchAt = switchAt
	d.retryAt = time.Time{}
	d.warned = false
}

// openLogFile создаёт каталог файла и открывает файл на дописывание.
func openLogFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), directoryPermission); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, permission)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	return file, nil
}
