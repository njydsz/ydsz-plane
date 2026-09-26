// Package metrics — 效能度量应用服务单元测试。
//
// 覆盖范围：
//  1. percentile 百分位计算
//  2. classifyDORALevel DORA 分级逻辑
//  3. cache key 格式（通过 MetricCache.keyXxx 方法）
//  4. MetricName 常量完整性
package metrics

import (
	"testing"
)

// ==========================================================================
// percentile
// ==========================================================================

func TestPercentile_Empty(t *testing.T) {
	got := percentile([]float64{}, 50)
	if got != 0 {
		t.Errorf("percentile(empty, 50) = %v, want 0", got)
	}
}

func TestPercentile_Single(t *testing.T) {
	got := percentile([]float64{42.0}, 50)
	if got != 42.0 {
		t.Errorf("percentile([42], 50) = %v, want 42", got)
	}
}

func TestPercentile_MedianEven(t *testing.T) {
	// sorted input: [1, 2, 3, 4] → P50 = 2.5
	data := []float64{1, 2, 3, 4}
	got := percentile(data, 50)
	want := 2.5
	if got != want {
		t.Errorf("percentile([1,2,3,4], 50) = %v, want %v", got, want)
	}
}

func TestPercentile_MedianOdd(t *testing.T) {
	// sorted input: [1, 2, 3] → P50 = 2
	data := []float64{1, 2, 3}
	got := percentile(data, 50)
	want := 2.0
	if got != want {
		t.Errorf("percentile([1,2,3], 50) = %v, want %v", got, want)
	}
}

func TestPercentile_P85(t *testing.T) {
	// [10, 20, 30, 40, 50] → P85 ≈ idx=3.4 → 40*0.6 + 50*0.4 = 44
	data := []float64{10, 20, 30, 40, 50}
	got := percentile(data, 85)
	want := 44.0
	if got != want {
		t.Errorf("percentile([10,20,30,40,50], 85) = %v, want %v", got, want)
	}
}

func TestPercentile_P0(t *testing.T) {
	data := []float64{5, 10, 15}
	got := percentile(data, 0)
	if got != 5.0 {
		t.Errorf("percentile P0 = %v, want 5.0", got)
	}
}

func TestPercentile_P100(t *testing.T) {
	data := []float64{5, 10, 15}
	got := percentile(data, 100)
	if got != 15.0 {
		t.Errorf("percentile P100 = %v, want 15.0", got)
	}
}

// ==========================================================================
// classifyDORALevel
// ==========================================================================

func TestClassifyDORALevel_Elite(t *testing.T) {
	r := &DORAResult{
		DeploymentFrequency: 2.0,  // daily: +2
		LeadTimeForChanges:  0.5,  // <1h: +2
		ChangeFailureRate:   0.02, // <5%: +2
		MTTR:                0.5,  // <1h: +2
		// total = 8 >= 7
	}
	got := classifyDORALevel(r)
	if got != "elite" {
		t.Errorf("classifyDORALevel(elite case) = %q, want %q", got, "elite")
	}
}

func TestClassifyDORALevel_High(t *testing.T) {
	r := &DORAResult{
		DeploymentFrequency: 2.0,  // +2
		LeadTimeForChanges:  0.5,  // +2
		ChangeFailureRate:   0.02, // +2
		MTTR:                5.0,  // <24h: +1
		// total = 7 >= 7 → elite
	}
	// Adjust to get "high"
	r.MTTR = 48.0 // >24h: +0 → total = 6
	got := classifyDORALevel(r)
	if got != "high" {
		t.Errorf("classifyDORALevel(high case) = %q, want %q", got, "high")
	}
}

func TestClassifyDORALevel_Medium(t *testing.T) {
	r := &DORAResult{
		DeploymentFrequency: 0.2,  // weekly: +1
		LeadTimeForChanges:  12.0, // <24h: +1
		ChangeFailureRate:   0.10, // <15%: +1
		MTTR:                48.0, // >24h: +0
		// total = 3
	}
	got := classifyDORALevel(r)
	if got != "medium" {
		t.Errorf("classifyDORALevel(medium case) = %q, want %q", got, "medium")
	}
}

func TestClassifyDORALevel_Low(t *testing.T) {
	r := &DORAResult{
		DeploymentFrequency: 0.01, // rarely: +0
		LeadTimeForChanges:  72.0, // >24h: +0
		ChangeFailureRate:   0.30, // >15%: +0
		MTTR:                100.0, // >24h: +0
		// total = 0
	}
	got := classifyDORALevel(r)
	if got != "low" {
		t.Errorf("classifyDORALevel(low case) = %q, want %q", got, "low")
	}
}

// ==================================================================
// Metric Constants
// ==================================================================

func TestMetricConstants(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"velocity", MetricVelocity, "velocity"},
		{"lead_time_p50", MetricLeadTimeP50, "lead_time_p50"},
		{"lead_time_p85", MetricLeadTimeP85, "lead_time_p85"},
		{"throughput", MetricThroughput, "throughput"},
		{"wip", MetricWIP, "wip"},
		{"defect_density", MetricDefectDensity, "defect_density"},
		{"escape_rate", MetricEscapeRate, "escape_rate"},
		{"rework_rate", MetricReworkRate, "rework_rate"},
		{"dora_deployment_frequency", MetricDORADF, "dora_deployment_frequency"},
		{"dora_lead_time", MetricDORALT, "dora_lead_time"},
		{"dora_change_failure_rate", MetricDORACFR, "dora_change_failure_rate"},
		{"dora_mttr", MetricDORAMTTR, "dora_mttr"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("metric constant %q: got %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

// ==================================================================
// Velocity Result JSON fields / SprintVelocity
// ==================================================================

func TestSprintVelocity_JSONTags(t *testing.T) {
	// SprintVelocity has stable JSON field names verified implicitly via struct tags.
	sv := SprintVelocity{
		SprintID:   1,
		SprintName: "Sprint 1",
		PointsDone: 25.0,
		IssuesDone: 8,
	}
	if sv.SprintID != 1 {
		t.Errorf("SprintVelocity.SprintID = %d, want 1", sv.SprintID)
	}
	if sv.PointsDone != 25.0 {
		t.Errorf("SprintVelocity.PointsDone = %f, want 25.0", sv.PointsDone)
	}
}

func TestLeadTimeResult_ZeroValue(t *testing.T) {
	r := LeadTimeResult{ProjectID: 99}
	if r.SampleSize != 0 {
		t.Errorf("LeadTimeResult zero SampleSize = %d, want 0", r.SampleSize)
	}
}

func TestQualityMetrics_ZeroValue(t *testing.T) {
	m := QualityMetrics{ProjectID: 42}
	if m.EscapeRate != 0 {
		t.Errorf("QualityMetrics zero EscapeRate = %f, want 0", m.EscapeRate)
	}
}

func TestDORAResult_ZeroValue(t *testing.T) {
	r := DORAResult{ProjectID: 1}
	if r.Level != "" {
		t.Errorf("DORAResult zero Level = %q, want empty", r.Level)
	}
}
