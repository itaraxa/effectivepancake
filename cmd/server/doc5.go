// Доработайте сервер, чтобы он мог изменять свои параметры запуска по умолчанию через переменные окружения:
//     ADDRESS отвечает за адрес эндпоинта HTTP-сервера.

// Приоритет параметров должен быть таким:
//     Если указана переменная окружения, то используется она.
//     Если нет переменной окружения, но есть аргумент командной строки (флаг), то используется он.
//     Если нет ни переменной окружения, ни флага, то используется значение по умолчанию.

package main
