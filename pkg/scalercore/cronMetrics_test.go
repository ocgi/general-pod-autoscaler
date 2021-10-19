// Copyright 2021 The OCGI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scalercore

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ocgi/general-pod-autoscaler/pkg/apis/autoscaling/v1alpha1"
)

type TestCronSchedule struct {
	name    string
	mode    v1alpha1.CronMetricMode
	desired int32
	gpa     *v1alpha1.GeneralPodAutoscaler
	time    time.Time
}

func TestInCronSchedule(t *testing.T) {
	var min1, max1 int32 = 5, 7
	var min2, max2 int32 = 6, 8
	var schedule1, schedule2 = "*/1 10-12 * * *", "*/1 9-10 * * *"
	testTime1, err := time.Parse("2006-01-02 15:04:05", "2020-12-18 09:04:41")
	if err != nil {
		t.Fatal(err)
	}
	gpa := &v1alpha1.GeneralPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Time{Time: testTime1.Add(-60 * time.Minute)},
		},
		Status: v1alpha1.GeneralPodAutoscalerStatus{},
	}

	tc := TestCronSchedule{
		name: "single timeRange, out of range",
		mode: v1alpha1.CronMetricMode{
			CronMetrics: []v1alpha1.CronMetricSpec{
				{
					Schedule:    schedule1,
					MinReplicas: &min1,
					MaxReplicas: max1,
				},
				{
					Schedule:    schedule2,
					MinReplicas: &min2,
					MaxReplicas: max2,
				},
			},
			DefaultReplicas: 5,
		},
	}
	t.Run(tc.name, func(t *testing.T) {
		defaultGPA := gpa
		if tc.gpa != nil {
			defaultGPA = tc.gpa
		}
		testTime := testTime1
		if !tc.time.IsZero() {
			testTime = tc.time
		}
		cron := &CronMetricsScaler{ranges: tc.mode.CronMetrics, name: Cron, now: testTime}
		actualMin, actualMax, schedule, _ := cron.GetCurrentMaxAndMinReplicas(defaultGPA)
		if actualMin != 6 || actualMax != 8 {
			t.Errorf("desired min: %v, max: %v, actual min: %v, max: %v", min2, max2, actualMin, actualMax)
		}
		if schedule != schedule2 {
			t.Errorf("desired schedule: %v, actual schedule: %v", schedule2, schedule)
		}
	})
}

func TestNotInCronSchedule(t *testing.T) {
	var min1, max1 int32 = 5, 7
	var min2, max2 int32 = 6, 8
	var schedule1, schedule2 = "*/1 10-12 * * *", "*/1 9-10 * * *"
	testTime1, err := time.Parse("2006-01-02 15:04:05", "2020-12-18 13:04:41")
	if err != nil {
		t.Fatal(err)
	}
	lastTime := metav1.Time{Time: testTime1.Add(-1 * time.Second)}
	gpa := &v1alpha1.GeneralPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Time{Time: testTime1.Add(-60 * time.Minute)},
		},
		Status: v1alpha1.GeneralPodAutoscalerStatus{
			LastCronScheduleTime: &lastTime,
		},
	}
	tc := TestCronSchedule{
		name: "single timeRange, out of range",
		mode: v1alpha1.CronMetricMode{
			CronMetrics: []v1alpha1.CronMetricSpec{
				{
					Schedule:    schedule1,
					MinReplicas: &min1,
					MaxReplicas: max1,
				},
				{
					Schedule:    schedule2,
					MinReplicas: &min2,
					MaxReplicas: max2,
				},
			},
			DefaultReplicas: 5,
		},
	}
	t.Run(tc.name, func(t *testing.T) {
		defaultGPA := gpa
		if tc.gpa != nil {
			defaultGPA = tc.gpa
		}
		testTime := testTime1
		if !tc.time.IsZero() {
			testTime = tc.time
		}
		cron := &CronMetricsScaler{ranges: tc.mode.CronMetrics, name: Cron, now: testTime}
		_, _, _, InCron := cron.GetCurrentMaxAndMinReplicas(defaultGPA)
		if InCron != false {
			t.Errorf("desired InCron: false, actual InCron: %v", InCron)
		}
	})
}

func TestAcrossPeriods(t *testing.T) {
	var min1, max1 int32 = 5, 7
	var min2, max2 int32 = 6, 8
	var schedule1, schedule2 = "0-59 10-12 * * *", "30-59 13-16 * * *"
	testTime1, err := time.Parse("2006-01-02 15:04:05", "2020-12-18 12:59:41")
	if err != nil {
		t.Fatal(err)
	}
	lastTime := metav1.Time{Time: testTime1.Add(-1 * time.Second)}
	gpa := &v1alpha1.GeneralPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Time{Time: testTime1.Add(-60 * time.Minute)},
		},
		Status: v1alpha1.GeneralPodAutoscalerStatus{
			LastCronScheduleTime: &lastTime,
		},
	}
	tc := TestCronSchedule{
		name: "single timeRange, out of range",
		mode: v1alpha1.CronMetricMode{
			CronMetrics: []v1alpha1.CronMetricSpec{
				{
					Schedule:    schedule1,
					MinReplicas: &min1,
					MaxReplicas: max1,
				},
				{
					Schedule:    schedule2,
					MinReplicas: &min2,
					MaxReplicas: max2,
				},
			},
			DefaultReplicas: 5,
		},
	}
	t.Run(tc.name, func(t *testing.T) {
		defaultGPA := gpa
		if tc.gpa != nil {
			defaultGPA = tc.gpa
		}
		testTime := testTime1
		if !tc.time.IsZero() {
			testTime = tc.time
		}
		cron := &CronMetricsScaler{ranges: tc.mode.CronMetrics, name: Cron, now: testTime}
		_, _, _, InCron := cron.GetCurrentMaxAndMinReplicas(defaultGPA)
		if InCron != false {
			t.Errorf("desired InCron: false, actual InCron: %v", InCron)
		}
	})
}
