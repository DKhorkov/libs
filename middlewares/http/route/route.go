package route

import (
	"net/http"
	"strings"
)

// Other — метка запроса без шаблона маршрута.
//
// Сырой путь в метку не идёт никогда: каждый идентификатор в пути стал бы
// отдельной серией, и число серий росло бы без предела.
const Other = "other"

// FromPattern — путь из шаблона стандартного ServeMux без метода и хоста.
//
// Шаблон ServeMux (Go 1.23+, поле Request.Pattern) имеет вид
// "[METHOD ][HOST]/[PATH]": "GET example.com/items/{id}" даёт "/items/{id}".
// Метод отрезается, иначе он задвоился бы с меткой method; хост — потому что
// маршрут один при любом хосте. Без шаблона — Other.
func FromPattern(r *http.Request) string {
	pattern := r.Pattern
	if pattern == "" {
		return Other
	}

	if i := strings.IndexAny(pattern, " \t"); i >= 0 {
		pattern = strings.TrimLeft(pattern[i:], " \t")
	}

	slash := strings.IndexByte(pattern, '/')
	if slash < 0 {
		return Other
	}

	return pattern[slash:]
}

// OtherMethod — метод вне стандартного набора.
const OtherMethod = "OTHER"

// knownMethods — методы, которые попадают в метку и имя спана как есть.
var knownMethods = map[string]struct{}{
	http.MethodGet: {}, http.MethodHead: {}, http.MethodPost: {},
	http.MethodPut: {}, http.MethodPatch: {}, http.MethodDelete: {},
	http.MethodOptions: {}, http.MethodConnect: {}, http.MethodTrace: {},
}

// Method — метод запроса для меток метрик и имени спана: стандартный — как
// есть, всё прочее — OtherMethod.
//
// Метод задаёт клиент, и gorilla пускает любой токен в маршрут без Methods
// и в обработчики 404/405. Без нормализации сканер, перебирающий методы,
// дал бы по серии метрик и по операции Jaeger на каждый.
func Method(r *http.Request) string {
	if _, ok := knownMethods[r.Method]; ok {
		return r.Method
	}

	return OtherMethod
}
