
Алгоритм поиска связных компонент, реализованный в параллельном режиме с помощью MPI.
В проекте несколько вариаций данного алгоритма.

## Установка
1. Скачать MPI с официального сайта [OpenMPI](https://www.open-mpi.org/)
2. Зайти в папку src
3. Запустить команду `go mod tidy`

## Запуск

1. Для сборки проекта выполнить файл build.sh (`./build.sh`)
2. Выполнить `./bin/main` с необходимыми флагами:

`--algo <string>` устанавливает, какой из алгоритмов будет запущен (`basic`, `fastsv`, `basic_mpi`, `fastsv_mpi`)

`--algo-variant <string>` указывает на вариант выбранного алгоритма (необходим для алгоритмов `basic_mpi`, `fastsv_mpi`)
* `basic_mpi`: `cont`
* `fastsv_mpi`: `disc_f_par`, `cont_f_par`, `cont_s_par`

`--file path/to/file` устанавливает путь к файлу, где находится граф

`--proc-num n` (необходим для алгоритмов `basic_mpi`, `fastsv_mpi`) указывает, на скольких процессах будут выполняться вычисления

`--logging` (необходим для алгоритмов `basic_mpi`, `fastsv_mpi`)
Наличие данного флага указывает, что будет выполнено логирование работы алгоритма

## Описание интерфейсов 

1. Интерфейс для чтения графа

Реализован в папке `src/graph`. Интерфейс GraphIO хранит 2 метода для чтения графа и записи результата.

В проекте представлена одна реализация данного интерфейса `FileGraphIO`. В ней граф читается из файла `.csv`. Также можно добавить поддержку других расширений, написав дополнительный метод.

**Далее все интерфейсы используются только в алгоритмах MPI. Они написаны в папке `src/algos`**

2. `IAlgo` необходим для классов Algo алгоритмов MPI. Каждый алгоритм реализует данный интерфейс. В нем есть методы Init и Close для инициализации и завершения алгоритма. Также есть метод GetLogger, который возвращает интерфейс логгера. И основной метод Run, который запускает работу алгоритма

3. `ILogger` необходим для логирования алгоритмов. В каждом алгоритме MPI есть объект класса Logger. Данные классы являются реализацией `ILogger`. В них есть методы Init, Close для инициализации и закрытия. Также для использования логгера необходимо запустить метод Start. После отрабатывания алгоритма вызвать метод Finish.

## Описание структуры проекта и принципа его работы

В проекте есть 2 точки входа `cmd/main` и `cmd/main_mpi`.

### `cmd/main`:

Главной точкой входа является `cmd/main`. В параметрах передается название алгоритма, который следует выполнить.
Для соответствующего алгоритма нужно выбрать адаптер:

* `adapter/adapter`: `basic`, `fastsv`

В данном адаптере происходит просто вызов функции алогритма и вывод результируещего вектора в консоль
* `adapter/adapter_mpi`: `basic_mpi`, `fastsv_mpi`

В данном адаптере идет запуск второй точки входа `cmd/main_mpi` с помощью команды `mpirun`. Параметрами командной строки также передаются название алгоритма и его конфигурация (где учитываются параметры командной строки для `cmd/main`). После запуска данной команды идет уже параллельная работа MPI узлов
Далее более подробное описание работы каждого узла

### `cmd/main_mpi`:

При запуске узла MPI происходит получение названия алгоритма из параметров командной строки и инициализация объекта алгоритма, который является реализацией интерфейса `IAlgo`:
```go
    // fh - объект для хранения значения всех флагов командной строки
    var algo ialgo.IAlgo = getAlgo(fh.Algo)
    err = algo.Init(conf)
    if err != nil {
        log.Panicln("algo Init:", err)
    }

```

И после этого происходит запуск алгоритма с логированием или без него, в зависимости от того, передан был флаг `logging` или нет.

### Работа алгоритмов MPI

В обоих алгоритмах используются структуры master для ведущего процесса и slave для ведомых процессов. 
Принцип работы функций Run в обоих алгоритмах похожий. Сначала идет распределение частей графа по ведомым процессам, затем вычисление результирующего вектора.

master контроллирует общую работу всех ведомых процессов и сообщается, в какой момент можно переходить на следующую стадию.

slave запускают 2 горутины для отправки и принятия любых сообщений при вычислении вектора. Таким образом, буфер обмена сообщениями MPI не будет переполняться.

#### Логирование

В необходимых точках алгоритма вставлены вызовы методов логгера, чтобы логгер понимал, на какой стадии сейчас находится алгоритм. В `basic_mpi/slave.go` внутри функции `CCSearch`:

```go
    slave.algo.logger.beginIterations()
	for {
		slave.algo.logger.beginIteration()
        ...
    }
```

Весь результат логирования попадает в папку корневой директории `log`, которая должна существовать. В ней создается папка, название которой - дата и время очередного запуска проекта. В ней лежат таблицы логирования каждого узла MPI и общий файл с выводом основной информации, такой как общее время работы.
