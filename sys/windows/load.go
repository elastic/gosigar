// +build windows

package windows

import (
	"sync"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/pkg/errors"
)

type win32_PerfFormattedData_PerfOS_System struct {
	ProcessorQueueLength int
}

type loadSample struct {
	timestamp time.Time
	value     int
}

type loadSamplesType struct {
	samples []loadSample
	lock    sync.RWMutex
}

const (
	LOAD_HISTORY_DURATION = 15 * time.Minute
)

var (
	// loadSamples keeps samples up to LOAD_HISTORY_DURATION old.
	loadSamples loadSamplesType
)

func init() {
	loadSamples = loadSamplesType{}
}

// GetLoadAverages returns the 1-minute, 5-minute, and 15-minute load averages for a Windows
// system. Load averages are based on processor queue lengths, which is the number of processes
// waiting for time on the CPU. This function also samples the current load average when called.
func GetLoadAverages() (float64, float64, float64, error) {
	// Sample current load
	currentLoad, err := getCurrentLoad()
	if err != nil {
		return 0, 0, 0, err
	}
	addLoadSample(currentLoad)

	// Calculate 1-minute, 5-minute, and 15-minute load averages
	one := getLoadAverage(1 * time.Duration)
	five := getLoadAverage(5 * time.Duration)
	fifteen := getLoadAverage(15 * time.Duration)

	return one, five, fifteen, nil
}

func getCurrentLoad() (int, error) {
	var dst []win32_PerfFormattedData_PerfOS_System
	q := wmi.CreateQuery(&dst, "")
	err := wmi.Query(q, &dst)
	if err != nil {
		return 0, errors.Wrap(err, "wmi query for Win32_PerfFormattedData_PerfOS_System failed")
	}
	if len(dst) != 1 {
		return 0, errors.New("wmi query for Win32_PerfFormattedData_PerfOS_System failed")
	}

	currentLoad := dst[0].ProcessorQueueLength
	return currentLoad
}

func addLoadSample(value int) {
	now := time.Now

	loadSamples.lock.Lock()
	defer loadSamples.lock.Unlock()

	loadSamples.samples = append(loadSamples.samples, value)

	// Cleanup old samples
	newLoadSamples = []loadSample{}
	for _, sample := range loadSamples.samples {
		if sample.timestamp.After(now.Add(-LOAD_HISTORY_DURATION)) {
			newLoadSamples = append(newLoadSamples, sample)
		}
	}
	loadSamples.samples = newLoadSamples
}

func getLoadAverage(avergeDuration time.Duration) float64 {
	loadSamples.lock.RLock()
	defer loadSamples.lock.RUnlock()

	total := 0
	count := 0
	startTime := now.Add(-avergeDuration)
	for _, sample := range loadSamples.samples {
		if sample.timestamp >= startTime {
			total += sample.value
			count++
		}
	}

	return float64(total / count)
}
