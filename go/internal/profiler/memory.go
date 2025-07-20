package profiler

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// MemoryProfiler tracks memory usage and allocation patterns
type MemoryProfiler struct {
	mu          sync.RWMutex
	snapshots   []MemorySnapshot
	startTime   time.Time
	isRecording bool
	sampleRate  time.Duration
	stopChan    chan bool
}

// MemorySnapshot represents a point-in-time memory measurement
type MemorySnapshot struct {
	Timestamp     time.Time
	Alloc         uint64  // Current allocated bytes
	TotalAlloc    uint64  // Cumulative allocated bytes
	Sys           uint64  // System memory obtained from OS
	Lookups       uint64  // Number of pointer lookups
	Mallocs       uint64  // Number of malloc calls
	Frees         uint64  // Number of free calls
	HeapAlloc     uint64  // Heap allocated bytes
	HeapSys       uint64  // Heap system bytes
	HeapIdle      uint64  // Heap idle bytes
	HeapInuse     uint64  // Heap in-use bytes
	HeapReleased  uint64  // Heap released bytes
	HeapObjects   uint64  // Number of heap objects
	StackInuse    uint64  // Stack in-use bytes
	StackSys      uint64  // Stack system bytes
	MSpanInuse    uint64  // MSpan in-use bytes
	MSpanSys      uint64  // MSpan system bytes
	MCacheInuse   uint64  // MCache in-use bytes
	MCacheSys     uint64  // MCache system bytes
	GCSys         uint64  // GC system bytes
	OtherSys      uint64  // Other system bytes
	NextGC        uint64  // Next GC target
	LastGC        uint64  // Last GC time (nanoseconds)
	PauseTotalNs  uint64  // Total GC pause time
	PauseNs       uint64  // Recent GC pause time
	NumGC         uint32  // Number of GC cycles
	NumForcedGC   uint32  // Number of forced GC cycles
	GCCPUFraction float64 // Fraction of CPU time spent in GC
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler(sampleRate time.Duration) *MemoryProfiler {
	return &MemoryProfiler{
		snapshots:  make([]MemorySnapshot, 0),
		sampleRate: sampleRate,
		stopChan:   make(chan bool, 1),
	}
}

// Start begins memory profiling
func (mp *MemoryProfiler) Start() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if mp.isRecording {
		return
	}

	mp.isRecording = true
	mp.startTime = time.Now()
	mp.snapshots = mp.snapshots[:0] // Clear previous snapshots

	go mp.recordingLoop()
}

// Stop ends memory profiling
func (mp *MemoryProfiler) Stop() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if !mp.isRecording {
		return
	}

	mp.isRecording = false
	mp.stopChan <- true
}

// recordingLoop continuously records memory snapshots
func (mp *MemoryProfiler) recordingLoop() {
	ticker := time.NewTicker(mp.sampleRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mp.takeSnapshot()
		case <-mp.stopChan:
			mp.takeSnapshot() // Take final snapshot
			return
		}
	}
}

// takeSnapshot captures current memory statistics
func (mp *MemoryProfiler) takeSnapshot() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	snapshot := MemorySnapshot{
		Timestamp:     time.Now(),
		Alloc:         m.Alloc,
		TotalAlloc:    m.TotalAlloc,
		Sys:           m.Sys,
		Lookups:       m.Lookups,
		Mallocs:       m.Mallocs,
		Frees:         m.Frees,
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapIdle:      m.HeapIdle,
		HeapInuse:     m.HeapInuse,
		HeapReleased:  m.HeapReleased,
		HeapObjects:   m.HeapObjects,
		StackInuse:    m.StackInuse,
		StackSys:      m.StackSys,
		MSpanInuse:    m.MSpanInuse,
		MSpanSys:      m.MSpanSys,
		MCacheInuse:   m.MCacheInuse,
		MCacheSys:     m.MCacheSys,
		GCSys:         m.GCSys,
		OtherSys:      m.OtherSys,
		NextGC:        m.NextGC,
		LastGC:        m.LastGC,
		PauseTotalNs:  m.PauseTotalNs,
		NumGC:         m.NumGC,
		NumForcedGC:   m.NumForcedGC,
		GCCPUFraction: m.GCCPUFraction,
	}

	// Get recent pause time
	if len(m.PauseNs) > 0 {
		snapshot.PauseNs = m.PauseNs[(m.NumGC+255)%256]
	}

	mp.mu.Lock()
	mp.snapshots = append(mp.snapshots, snapshot)
	mp.mu.Unlock()
}

// GetSnapshots returns all recorded memory snapshots
func (mp *MemoryProfiler) GetSnapshots() []MemorySnapshot {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	// Return a copy to prevent race conditions
	snapshots := make([]MemorySnapshot, len(mp.snapshots))
	copy(snapshots, mp.snapshots)
	return snapshots
}

// GetSummary returns a summary of memory usage
func (mp *MemoryProfiler) GetSummary() MemorySummary {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	if len(mp.snapshots) == 0 {
		return MemorySummary{}
	}

	first := mp.snapshots[0]
	last := mp.snapshots[len(mp.snapshots)-1]

	var maxAlloc, maxHeapAlloc, maxSys uint64
	var totalGCPauses uint64
	var maxGCPause uint64

	for _, snapshot := range mp.snapshots {
		if snapshot.Alloc > maxAlloc {
			maxAlloc = snapshot.Alloc
		}
		if snapshot.HeapAlloc > maxHeapAlloc {
			maxHeapAlloc = snapshot.HeapAlloc
		}
		if snapshot.Sys > maxSys {
			maxSys = snapshot.Sys
		}
		if snapshot.PauseNs > maxGCPause {
			maxGCPause = snapshot.PauseNs
		}
	}

	totalGCPauses = last.PauseTotalNs - first.PauseTotalNs

	return MemorySummary{
		Duration:         last.Timestamp.Sub(first.Timestamp),
		StartAlloc:       first.Alloc,
		EndAlloc:         last.Alloc,
		MaxAlloc:         maxAlloc,
		StartHeapAlloc:   first.HeapAlloc,
		EndHeapAlloc:     last.HeapAlloc,
		MaxHeapAlloc:     maxHeapAlloc,
		StartSys:         first.Sys,
		EndSys:           last.Sys,
		MaxSys:           maxSys,
		TotalAllocations: last.TotalAlloc - first.TotalAlloc,
		NetAllocations:   last.Mallocs - first.Mallocs - (last.Frees - first.Frees),
		GCCycles:         last.NumGC - first.NumGC,
		TotalGCPauseTime: totalGCPauses,
		MaxGCPauseTime:   maxGCPause,
		AvgGCPauseTime:   totalGCPauses / uint64(max(1, last.NumGC-first.NumGC)),
		GCCPUFraction:    last.GCCPUFraction,
		SnapshotCount:    len(mp.snapshots),
	}
}

// MemorySummary provides aggregated memory usage statistics
type MemorySummary struct {
	Duration         time.Duration
	StartAlloc       uint64
	EndAlloc         uint64
	MaxAlloc         uint64
	StartHeapAlloc   uint64
	EndHeapAlloc     uint64
	MaxHeapAlloc     uint64
	StartSys         uint64
	EndSys           uint64
	MaxSys           uint64
	TotalAllocations uint64
	NetAllocations   uint64
	GCCycles         uint32
	TotalGCPauseTime uint64
	MaxGCPauseTime   uint64
	AvgGCPauseTime   uint64
	GCCPUFraction    float64
	SnapshotCount    int
}

// String returns a formatted string representation of the memory summary
func (ms MemorySummary) String() string {
	return fmt.Sprintf(`Memory Profile Summary:
Duration: %v
Allocated Memory: %s -> %s (max: %s)
Heap Memory: %s -> %s (max: %s)
System Memory: %s -> %s (max: %s)
Total Allocations: %s
Net Allocations: %d
GC Cycles: %d
GC Pause Time: total=%v, max=%v, avg=%v
GC CPU Fraction: %.2f%%
Snapshots: %d`,
		ms.Duration,
		formatBytes(ms.StartAlloc), formatBytes(ms.EndAlloc), formatBytes(ms.MaxAlloc),
		formatBytes(ms.StartHeapAlloc), formatBytes(ms.EndHeapAlloc), formatBytes(ms.MaxHeapAlloc),
		formatBytes(ms.StartSys), formatBytes(ms.EndSys), formatBytes(ms.MaxSys),
		formatBytes(ms.TotalAllocations),
		ms.NetAllocations,
		ms.GCCycles,
		time.Duration(ms.TotalGCPauseTime), time.Duration(ms.MaxGCPauseTime), time.Duration(ms.AvgGCPauseTime),
		ms.GCCPUFraction*100,
		ms.SnapshotCount)
}

// formatBytes formats byte counts in human-readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// max returns the maximum of two uint32 values
func max(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

// ProfiledOperation wraps an operation with memory profiling
func ProfiledOperation(name string, operation func() error) error {
	profiler := NewMemoryProfiler(10 * time.Millisecond)

	fmt.Printf("Starting profiled operation: %s\n", name)
	profiler.Start()

	err := operation()

	profiler.Stop()
	summary := profiler.GetSummary()

	fmt.Printf("Completed profiled operation: %s\n", name)
	fmt.Println(summary.String())

	return err
}

// AllocationTracker tracks allocations for specific operations
type AllocationTracker struct {
	name       string
	startStats runtime.MemStats
	endStats   runtime.MemStats
}

// NewAllocationTracker creates a new allocation tracker
func NewAllocationTracker(name string) *AllocationTracker {
	tracker := &AllocationTracker{name: name}
	runtime.GC() // Force GC for accurate measurement
	runtime.ReadMemStats(&tracker.startStats)
	return tracker
}

// Stop stops tracking and returns allocation statistics
func (at *AllocationTracker) Stop() AllocationStats {
	runtime.GC() // Force GC for accurate measurement
	runtime.ReadMemStats(&at.endStats)

	return AllocationStats{
		Name:             at.name,
		AllocatedBytes:   at.endStats.TotalAlloc - at.startStats.TotalAlloc,
		AllocatedObjects: at.endStats.Mallocs - at.startStats.Mallocs,
		FreedObjects:     at.endStats.Frees - at.startStats.Frees,
		NetObjects:       (at.endStats.Mallocs - at.startStats.Mallocs) - (at.endStats.Frees - at.startStats.Frees),
		HeapGrowth:       int64(at.endStats.HeapAlloc) - int64(at.startStats.HeapAlloc),
		GCCycles:         at.endStats.NumGC - at.startStats.NumGC,
	}
}

// AllocationStats contains allocation statistics for an operation
type AllocationStats struct {
	Name             string
	AllocatedBytes   uint64
	AllocatedObjects uint64
	FreedObjects     uint64
	NetObjects       uint64
	HeapGrowth       int64
	GCCycles         uint32
}

// String returns a formatted string representation of allocation stats
func (as AllocationStats) String() string {
	return fmt.Sprintf(`Allocation Stats for %s:
Allocated: %s (%d objects)
Freed: %d objects
Net Objects: %d
Heap Growth: %s
GC Cycles: %d`,
		as.Name,
		formatBytes(as.AllocatedBytes), as.AllocatedObjects,
		as.FreedObjects,
		as.NetObjects,
		formatBytes(uint64(max64(0, as.HeapGrowth))),
		as.GCCycles)
}

// max64 returns the maximum of two int64 values
func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
