// Задание по треку «Сервис сбора метрик и алертинга»
// Разработайте агент (HTTP-клиент) для сбора рантайм-метрик и их последующей отправки на сервер по протоколу HTTP.
// Агент должен собирать метрики двух типов:

//     Тип gauge, float64.
//     Тип counter, int64.

// В качестве источника метрик используйте пакет runtime.
// Нужно собирать следующие метрики типа gauge:

//     Alloc
//     BuckHashSys
//     Frees
//     GCCPUFraction
//     GCSys
//     HeapAlloc
//     HeapIdle
//     HeapInuse
//     HeapObjects
//     HeapReleased
//     HeapSys
//     LastGC
//     Lookups
//     MCacheInuse
//     MCacheSys
//     MSpanInuse
//     MSpanSys
//     Mallocs
//     NextGC
//     NumForcedGC
//     NumGC
//     OtherSys
//     PauseTotalNs
//     StackInuse
//     StackSys
//     Sys
//     TotalAlloc

// К метрикам пакета runtime добавьте ещё две:

//     PollCount (тип counter) — счётчик, увеличивающийся на 1 при каждом обновлении метрики из пакета runtime (на каждый pollInterval — см. ниже).
//     RandomValue (тип gauge) — обновляемое произвольное значение.

// По умолчанию приложение должно:

//     Обновлять метрики из пакета runtime с заданной частотой: pollInterval — 2 секунды.
//     Отправлять метрики на сервер с заданной частотой: reportInterval — 10 секунд.

// Чтобы приостанавливать работу функции на заданное время, используйте вызов time.Sleep(n * time.Second). Подробнее о пакете time и его возможностях вы узнаете в третьем спринте.
// Метрики нужно отправлять по протоколу HTTP методом POST:

//     Формат данных — http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>.
//     Адрес сервера — http://localhost:8080.
//     Заголовок — Content-Type: text/plain.

// Пример запроса к серверу:

// POST /update/counter/someMetric/527 HTTP/1.1
// Host: localhost:8080
// Content-Length: 0
// Content-Type: text/plain

// Пример ответа от сервера:

// HTTP/1.1 200 OK
// Date: Tue, 21 Feb 2023 02:51:35 GMT
// Content-Length: 11
// Content-Type: text/plain; charset=utf-8

// Покройте код агента и сервера юнит-тестами.

package main
