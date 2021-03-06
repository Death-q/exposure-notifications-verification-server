// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package issueapi

import (
	enobservability "github.com/google/exposure-notifications-server/pkg/observability"
	"github.com/google/exposure-notifications-verification-server/pkg/observability"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

const metricPrefix = observability.MetricRoot + "/api/issue"

var (
	mRealmTokenRemaining = stats.Int64(metricPrefix+"/realm_token_remaining", "Remaining number of verification codes", stats.UnitDimensionless)
	mRealmTokenCapacity  = stats.Float64(metricPrefix+"/realm_token_capacity", "Capacity utilization for issuing verification codes", stats.UnitDimensionless)

	mRequest = stats.Int64(metricPrefix+"/request", "# of code issue requests", stats.UnitDimensionless)

	mSMSRequest = stats.Int64(metricPrefix+"/sms_request", "# of sms requests", stats.UnitDimensionless)

	mRealmToken = stats.Int64(metricPrefix+"/realm_token", "# of realm tokens from limiter", stats.UnitDimensionless)

	mRealmTokenUsed = stats.Int64(metricPrefix+"/realm_token_used", "# of realm token used.", stats.UnitDimensionless)
)

var (
	// tokenStateTagKey indicate the state of the tokens. It's either "USED" or
	// "AVAILABLE".
	tokenStateTagKey = tag.MustNewKey("state")
)

func tokenAvailableTag() tag.Mutator {
	return tag.Upsert(tokenStateTagKey, "AVAILABLE")
}

func tokenLimitTag() tag.Mutator {
	return tag.Upsert(tokenStateTagKey, "LIMIT")
}

func init() {
	enobservability.CollectViews([]*view.View{
		{
			Name:        metricPrefix + "/realm_token_remaining_latest",
			Description: "Latest realm remaining tokens",
			TagKeys:     observability.CommonTagKeys(),
			Measure:     mRealmTokenRemaining,
			Aggregation: view.LastValue(),
		}, {
			Name:        metricPrefix + "/realm_token_capacity_latest",
			Description: "Latest realm token capacity utilization",
			TagKeys:     observability.CommonTagKeys(),
			Measure:     mRealmTokenCapacity,
			Aggregation: view.LastValue(),
		}, {
			Name:        metricPrefix + "/request_count",
			Measure:     mRequest,
			Description: "Count of code issue requests",
			TagKeys:     observability.APITagKeys(),
			Aggregation: view.Count(),
		}, {
			Name:        metricPrefix + "/sms_request_count",
			Measure:     mSMSRequest,
			Description: "The # of SMS requests",
			TagKeys:     append(observability.CommonTagKeys(), observability.ResultTagKey),
			Aggregation: view.Count(),
		}, {
			Name:        metricPrefix + "/realm_token_latest",
			Description: "Latest realm token count",
			TagKeys:     append(observability.CommonTagKeys(), tokenStateTagKey),
			Measure:     mRealmToken,
			Aggregation: view.LastValue(),
		}, {
			Name:        metricPrefix + "/realm_token_used_count",
			Description: "The count of # of realm token used.",
			TagKeys:     observability.CommonTagKeys(),
			Measure:     mRealmTokenUsed,
			Aggregation: view.Count(),
		},
	}...)
}
