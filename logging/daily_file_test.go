package logging

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestDailyFileSwitchesAtMidnight — запись после полуночи уходит в файл новой
// даты, а файл прошлого дня перестаёт расти.
//
// Раньше путь считался один раз на старте, и файл одного запуска рос до
// перезапуска: удаление по возрасту (mtime) его не находило — в него писали.
// Дата — в часовом поясе приложения, а не в UTC: в Москве полночь наступает
// в 21:00 UTC.
func TestDailyFileSwitchesAtMidnight(t *testing.T) {
	t.Parallel()

	moscow := time.FixedZone("MSK", 3*60*60)
	directory := filepath.Join(t.TempDir(), "logs")
	now := time.Date(2026, 10, 13, 20, 59, 50, 0, time.UTC) // 23:59:50 MSK

	file := NewDailyFile(filepath.Join(directory, "%s.log"), "2006-01-02", moscow)
	file.now = func() time.Time { return now }

	_, err := file.Write([]byte("before midnight\n"))
	require.NoError(t, err)

	now = now.Add(15 * time.Second) // 00:00:05 MSK следующего дня

	_, err = file.Write([]byte("after midnight\n"))
	require.NoError(t, err)
	require.NoError(t, file.Close())

	previous, err := os.ReadFile(filepath.Join(directory, "2026-10-13.log"))
	require.NoError(t, err, "каталог логов создаётся сам")
	require.Equal(t, "before midnight\n", string(previous))

	next, err := os.ReadFile(filepath.Join(directory, "2026-10-14.log"))
	require.NoError(t, err)
	require.Equal(t, "after midnight\n", string(next))
}

// TestDailyFileAppendsWithinDay — перезапуск в тот же день дописывает файл
// дня, а не затирает его.
func TestDailyFileAppendsWithinDay(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	pattern := filepath.Join(directory, "%s.log")
	now := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)

	for _, line := range []string{"first run\n", "second run\n"} {
		file := NewDailyFile(pattern, "2006-01-02", time.UTC)
		file.now = func() time.Time { return now }

		_, err := file.Write([]byte(line))
		require.NoError(t, err)
		require.NoError(t, file.Close())
	}

	content, err := os.ReadFile(filepath.Join(directory, "2026-10-13.log"))
	require.NoError(t, err)
	require.Equal(t, "first run\nsecond run\n", string(content))
}

// TestDailyFileIsSafeForConcurrentWrites — логгер пишет из многих горутин, и
// строки не теряются и не перемешиваются.
func TestDailyFileIsSafeForConcurrentWrites(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	file := NewDailyFile(filepath.Join(directory, "%s.log"), "2006-01-02", time.UTC)

	const writers = 50

	var group sync.WaitGroup

	for range writers {
		group.Go(func() {
			_, err := file.Write([]byte("line\n"))
			require.NoError(t, err)
		})
	}

	group.Wait()
	require.NoError(t, file.Close())

	matches, err := filepath.Glob(filepath.Join(directory, "*.log"))
	require.NoError(t, err)
	require.Len(t, matches, 1)

	content, err := os.ReadFile(matches[0])
	require.NoError(t, err)
	require.Len(t, content, writers*len("line\n"))
}

// TestDailyFileWarnsOnceAndBacksOff — недоступный файл даёт одно
// предупреждение, а повторная попытка открыть — не чаще раза в минуту.
//
// Без этого сбой молчал (в отличие от прежнего New), а каждая строка лога
// заново пыталась создать каталог и открыть файл под общим мьютексом — и
// тормозила весь логгер.
func TestDailyFileWarnsOnceAndBacksOff(t *testing.T) {
	t.Parallel()

	var (
		warnings bytes.Buffer
		attempts int
	)

	now := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)

	file := NewDailyFile("unwritable/%s.log", "2006-01-02", time.UTC)
	file.now = func() time.Time { return now }
	file.warnings = &warnings
	file.openFile = func(string) (*os.File, error) {
		attempts++

		return nil, errors.New("permission denied")
	}

	for range 3 {
		_, err := file.Write([]byte("line\n"))
		require.Error(t, err)
	}

	require.Equal(t, 1, attempts, "повтор раньше минуты")
	require.Equal(t, 1, strings.Count(warnings.String(), "unwritable/2026-10-13.log"))

	now = now.Add(retryInterval)

	_, err := file.Write([]byte("line\n"))
	require.Error(t, err)
	require.Equal(t, 2, attempts, "через минуту пробует снова")
	require.Equal(t, 1, strings.Count(warnings.String(), "Failed to open log file"),
		"предупреждение — одно на сбой, а не на каждую попытку")
}

// TestDailyFileWithoutLocationUsesUTC — часовой пояс nil считается как UTC,
// а не роняет логгер паникой time.Time.In(nil).
func TestDailyFileWithoutLocationUsesUTC(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	file := NewDailyFile(filepath.Join(directory, "%s.log"), "2006-01-02", nil)
	file.now = func() time.Time { return time.Date(2026, 10, 13, 23, 30, 0, 0, time.UTC) }

	_, err := file.Write([]byte("line\n"))
	require.NoError(t, err)
	require.NoError(t, file.Close())
	require.FileExists(t, filepath.Join(directory, "2026-10-13.log"))
}
