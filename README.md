# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Результат профилирования потребления памяти
Что исправлено:
1. Заменил fmt.Sprintf("%s/%s", …) на более дешевую конкатенацию строк base + "/" + id.
2. Предварительная аллокация памяти, вместо растущего слайса - result := make([]byte, 0, 11) в encodeBase62.

Showing nodes accounting for -13883.49kB, 31.64% of 43876.79kB total
Dropped 9 nodes (cum <= 219.38kB)
flat  flat%   sum%        cum   cum%
-11733.62kB 26.74% 26.74% -13371.80kB 30.48%  compress/flate.NewWriter (inline)
-1638.18kB  3.73% 30.48% -1638.18kB  3.73%  compress/flate.(*compressor).initDeflate (inline)
1024.44kB  2.33% 28.14%  1024.44kB  2.33%  runtime.malg
514kB  1.17% 26.97%      514kB  1.17%  bufio.NewReaderSize (inline)
-513.69kB  1.17% 28.14%  -513.59kB  1.17%  Ustasjs/yp-url-shortener/internal/service/shortener.(*Shortener).CreateShortURLsBatch
-512.25kB  1.17% 29.31%  -512.25kB  1.17%  go.uber.org/zap/internal/stacktrace.init.func1
-512.17kB  1.17% 30.48%  -512.17kB  1.17%  net/textproto.readMIMEHeader
512.16kB  1.17% 29.31%   512.16kB  1.17%  github.com/jackc/pgx/v5.(*ConnConfig).Copy (inline)
-512.12kB  1.17% 30.47%  -512.12kB  1.17%  net/http.ListenAndServe (inline)
-512.05kB  1.17% 31.64%  -512.05kB  1.17%  context.(*cancelCtx).Done
0     0% 31.64% -13885.38kB 31.65%  Ustasjs/yp-url-shortener/internal/handler.(*Handler).CreateShortURLSByBatch
0     0% 31.64% -14397.63kB 32.81%  Ustasjs/yp-url-shortener/internal/logger.LoggerMiddleware.func1
0     0% 31.64% -13885.38kB 31.65%  Ustasjs/yp-url-shortener/internal/middleware.GzipDecompress.func1
0     0% 31.64% -13885.38kB 31.65%  Ustasjs/yp-url-shortener/internal/router.initMiddleware.Auth.func1.1
0     0% 31.64%      514kB  1.17%  bufio.NewReader (inline)
0     0% 31.64% -1638.18kB  3.73%  compress/flate.(*compressor).init
0     0% 31.64% -13371.80kB 30.48%  compress/gzip.(*Writer).Write
0     0% 31.64% -13371.80kB 30.48%  encoding/json.(*Encoder).Encode
0     0% 31.64% -14397.63kB 32.81%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
0     0% 31.64% -13885.38kB 31.65%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
0     0% 31.64% -13885.38kB 31.65%  github.com/go-chi/chi/v5/middleware.(*Compressor).Handler-fm.(*Compressor).Handler.func1
0     0% 31.64% -13371.80kB 30.48%  github.com/go-chi/chi/v5/middleware.(*compressResponseWriter).Write
0     0% 31.64%   512.16kB  1.17%  github.com/jackc/pgx/v5.ConnectConfig
0     0% 31.64%   512.16kB  1.17%  github.com/jackc/pgx/v5/stdlib.(*driverConnector).Connect
0     0% 31.64%  -512.25kB  1.17%  go.uber.org/zap.(*Logger).Info
0     0% 31.64%  -512.25kB  1.17%  go.uber.org/zap.(*Logger).check
0     0% 31.64%  -512.25kB  1.17%  go.uber.org/zap/internal/pool.(*Pool[go.shape.*uint8]).Get (inline)
0     0% 31.64%  -512.25kB  1.17%  go.uber.org/zap/internal/stacktrace.Capture
0     0% 31.64%  -512.25kB  1.17%  go.uber.org/zap/internal/stacktrace.init.New[go.shape.*uint8].func2
0     0% 31.64%  -512.12kB  1.17%  main.main.func1
0     0% 31.64% -14396.80kB 32.81%  net/http.(*conn).serve
0     0% 31.64% -14397.63kB 32.81%  net/http.HandlerFunc.ServeHTTP
0     0% 31.64%      514kB  1.17%  net/http.newBufioReader
0     0% 31.64%  -512.17kB  1.17%  net/http.readRequest
0     0% 31.64% -14397.63kB 32.81%  net/http.serverHandler.ServeHTTP
0     0% 31.64%  -512.17kB  1.17%  net/textproto.(*Reader).ReadMIMEHeader (inline)
0     0% 31.64%     -513kB  1.17%  runtime.mcall
0     0% 31.64%      513kB  1.17%  runtime.mstart
0     0% 31.64%      513kB  1.17%  runtime.mstart0
0     0% 31.64%      513kB  1.17%  runtime.mstart1
0     0% 31.64%  1024.44kB  2.33%  runtime.newproc.func1
0     0% 31.64%  1024.44kB  2.33%  runtime.newproc1
0     0% 31.64%     -513kB  1.17%  runtime.park_m
0     0% 31.64%  1024.44kB  2.33%  runtime.systemstack
0     0% 31.64%  -512.25kB  1.17%  sync.(*Pool).Get