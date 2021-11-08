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
	"time"

	"github.com/robfig/cron"
	"k8s.io/klog"

	"github.com/ocgi/general-pod-autoscaler/pkg/apis/autoscaling/v1alpha1"
)

var _ Scaler = &CronMetricsScaler{}
var recordCronMetricsScheduleName = ""

// CronScaler is a crontab GPA
type CronMetricsScaler struct {
	ranges     []v1alpha1.CronMetricSpec
	defaultSet v1alpha1.CronMetricSpec
	name       string
	now        time.Time
}

// NewCronScaler initializer crontab GPA
func NewCronMetricsScaler(ranges []v1alpha1.CronMetricSpec) *CronMetricsScaler {
	var def v1alpha1.CronMetricSpec
	filter := make([]v1alpha1.CronMetricSpec, 0)
	for _, cr := range ranges {
		if cr.Schedule != "default" {
			filter = append(filter, cr)
		} else {
			def = cr
		}
	}
	return &CronMetricsScaler{ranges: filter, name: Cron, now: time.Now(), defaultSet: def}
}

// GetReplicas return replicas  recommend by crontab GPA
func (s *CronMetricsScaler) GetReplicas(gpa *v1alpha1.GeneralPodAutoscaler, currentReplicas int32) (int32, error) {
	var max int32 = 0
	for _, t := range s.ranges {
		misMatch, finalMatch, err := s.getFinalMatchAndMisMatch(gpa, t.Schedule)
		if err != nil {
			klog.Error(err)
			return currentReplicas, nil
		}
		klog.Infof("firstMisMatch: %v, finalMatch: %v", misMatch, finalMatch)
		if finalMatch == nil {
			continue
		}
		if max < t.MaxReplicas {
			max = t.MaxReplicas
			recordCronMetricsScheduleName = t.Schedule
		}
		klog.Infof("Schedule %v recommend %v replicas, desire: %v", t.Schedule, max, t.MaxReplicas)
	}
	if max == 0 {
		klog.Info("Recommend 0 replicas, use current replicas number")
		max = gpa.Status.DesiredReplicas
	}
	return max, nil
}

// get current cron config max and min replicas
func (s *CronMetricsScaler) GetCurrentMaxAndMinReplicas(gpa *v1alpha1.GeneralPodAutoscaler) (int32, int32, string) {
	var max, min int32
	//use defaultSet max min replicas
	max = s.defaultSet.MaxReplicas
	min = *s.defaultSet.MinReplicas
	recordCronMetricsScheduleName = s.defaultSet.Schedule
	//only one schedule satisfy
	for _, cr := range s.ranges {
		if cr.Schedule == "default" {
			//ignore `default` cron set
			continue
		}
		misMatch, finalMatch, err := s.getFinalMatchAndMisMatch(gpa, cr.Schedule)
		if err != nil {
			klog.Error(err)
			return max, min, recordCronMetricsScheduleName
		}
		klog.Infof("firstMisMatch: %v, finalMatch: %v, schedule: %v", misMatch, finalMatch, cr.Schedule)
		if finalMatch == nil {
			continue
		} else {
			max = cr.MaxReplicas
			min = *cr.MinReplicas
			recordCronMetricsScheduleName = cr.Schedule
			klog.Infof("Schedule %v recommend %v max replicas, min replicas: %v", cr.Schedule, max, min)
			return max, min, recordCronMetricsScheduleName
		}
	}
	return max, min, recordCronMetricsScheduleName
}

// ScalerName returns scaler name
func (s *CronMetricsScaler) ScalerName() string {
	return s.name
}

func (s *CronMetricsScaler) getFinalMatchAndMisMatch(gpa *v1alpha1.GeneralPodAutoscaler, schedule string) (*time.Time, *time.Time, error) {
	sched, err := cron.ParseStandard(schedule)
	if err != nil {
		return nil, nil, err
	}
	lastTime := gpa.Status.LastCronScheduleTime.DeepCopy()
	if recordCronMetricsScheduleName != schedule {
		lastTime = nil
	}
	if lastTime == nil || lastTime.IsZero() {
		lastTime = gpa.CreationTimestamp.DeepCopy()
	}
	match := lastTime.Time
	misMatch := lastTime.Time
	klog.Infof("Init time: %v, now: %v", lastTime, s.now)
	t := lastTime.Time
	for {
		if !t.After(s.now) {
			misMatch = t
			t = sched.Next(t)
			continue
		}
		match = t
		break
	}
	// fix bug: misMatch diff s.now < 1 ,but match diff s.now > 1
	if s.now.Sub(misMatch).Minutes() < 1 && s.now.After(misMatch) && match.Sub(s.now).Minutes() < 1 {
		return &misMatch, &match, nil
	}

	return nil, nil, nil
}
