package packets

import (
	"sort"
	"sync"
	"sync/atomic"
)

const (
	minBitSize = 6  // 2^6=64 соответствует размеру кэш-линии процессора
	steps      = 20 // Количество шагов для размеров буферов

	minSize = 1 << minBitSize // Минимальный размер буфера
	// Максимальный размер вычисляется динамически во время калибровки

	calibrateCallsThreshold = 42000 // Порог вызовов для активации калибровки
	maxPercentile           = 0.95  // Перцентиль для определения максимального размера
)

// pools представляет пул байтовых буферов.
//
// Разные пулы могут использоваться для разных типов буферов.
// Правильно определенные типы буферов со своими пулами помогают уменьшить
// потребление памяти.
type pools struct {
	calls [steps]uint64 // Счетчики вызовов для каждого размера буфера

	calibrating uint64 // Флаг выполнения калибровки (0 или 1)

	defaultSize uint64 // Размер по умолчанию для новых буферов
	maxSize     uint64 // Максимальный допустимый размер для возврата в пул

	pool sync.Pool // Основной пул для хранения буферов
}

var pool pools // Глобальный экземпляр пула

// GetBuffer возвращает пустой буфер из пула.
//
// После использования буфер должен быть возвращен в пул через PutBuffer.
// Это уменьшает количество аллокаций памяти для управления буферами.
func GetBuffer() *Buffer { return pool.get() }

// get возвращает новый буфер с длиной 0.
//
// Буфер может быть возвращен в пул через Put для снижения нагрузки на GC.
func (p *pools) get() *Buffer {
	v := p.pool.Get()
	if v != nil {
		return v.(*Buffer)
	}
	return &Buffer{
		b: make([]byte, 0, atomic.LoadUint64(&p.defaultSize)),
	}
}

// Put  возвращает буфер в пул.
//
// Буфер не должен использоваться после возврата в пул во избежание
// состояний гонки.
func Put(b *Buffer) { pool.put(b) }

// put возвращает буфер в пул.
//
// После возврата в пул к буферу нельзя обращаться.
func (p *pools) put(b *Buffer) {
	// Определяем индекс размера в массиве calls
	idx := index(len(b.b))

	// Если превышен порог вызовов - запускаем калибровку
	if atomic.AddUint64(&p.calls[idx], 1) > calibrateCallsThreshold {
		p.calibrate()
	}

	// Проверяем максимальный допустимый размер
	maxSize := int(atomic.LoadUint64(&p.maxSize))
	if maxSize == 0 || cap(b.b) <= maxSize {
		b.Reset()
		p.pool.Put(b)
	}
}

// calibrate настраивает оптимальные размеры буферов на основе статистики.
func (p *pools) calibrate() {
	// Захватываем флаг калибровки
	if !atomic.CompareAndSwapUint64(&p.calibrating, 0, 1) {
		return
	}

	// Собираем статистику использования
	a := make(callSizes, 0, steps)
	var callsSum uint64
	for i := uint64(0); i < steps; i++ {
		calls := atomic.SwapUint64(&p.calls[i], 0)
		callsSum += calls
		a = append(a, callSize{
			calls: calls,
			size:  minSize << i,
		})
	}

	// Сортируем размеры по убыванию частоты использования
	sort.Sort(a)

	// Вычисляем размер по умолчанию и максимальный размер
	defaultSize := a[0].size
	maxSize := defaultSize

	// Определяем размер, покрывающий 95% запросов
	maxSum := uint64(float64(callsSum) * maxPercentile)
	callsSum = 0
	for i := 0; i < steps; i++ {
		if callsSum > maxSum {
			break
		}
		callsSum += a[i].calls
		size := a[i].size
		if size > maxSize {
			maxSize = size
		}
	}

	// Обновляем настройки пула
	atomic.StoreUint64(&p.defaultSize, defaultSize)
	atomic.StoreUint64(&p.maxSize, maxSize)

	// Сбрасываем флаг калибровки
	atomic.StoreUint64(&p.calibrating, 0)
}

// callSize представляет связь между количеством вызовов и размером буфера
type callSize struct {
	calls uint64 // Количество запросов
	size  uint64 // Размер буфера
}

// callSizes реализует интерфейс сортировки для среза callSize
type callSizes []callSize

func (ci callSizes) Len() int           { return len(ci) }
func (ci callSizes) Less(i, j int) bool { return ci[i].calls > ci[j].calls }
func (ci callSizes) Swap(i, j int)      { ci[i], ci[j] = ci[j], ci[i] }

// index вычисляет индекс в массиве calls для заданного размера буфера
func index(n int) int {
	n--
	n >>= minBitSize
	idx := 0
	for n > 0 {
		n >>= 1
		idx++
	}
	// Ограничиваем максимальный индекс
	if idx >= steps {
		idx = steps - 1
	}
	return idx
}
