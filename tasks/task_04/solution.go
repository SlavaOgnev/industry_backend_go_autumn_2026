package main

type Stats struct {
	Count         int
	Sum, Min, Max int64
}

func Calc(nums []int64) Stats {
	if len(nums) < 2 {
		return Stats{}
	}
	first := nums[1] - nums[0]
	stats := Stats{Sum: 0, Count: 0, Min: first, Max: first}

	for i := 1; i < len(nums); i++ {
		define := nums[i] - nums[i-1]
		stats.Count++
		stats.Sum += define
		stats.Min = min(stats.Min, define)
		stats.Max = max(stats.Max, define)
	}

	return stats

}
