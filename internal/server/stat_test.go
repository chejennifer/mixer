// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestFilterAndRank(t *testing.T) {
	for _, c := range []struct {
		input   *ObsTimeSeries
		mmethod string
		unit    string
		op      string
		want    *ObsTimeSeries
	}{
		// Default ranking
		{
			&ObsTimeSeries{
				SourceSeries: []*SourceSeries{
					{
						Val:               map[string]float64{"2011": 100, "2012": 101},
						MeasurementMethod: "CensusPEPSurvey",
						ImportName:        "CensusPEP",
						ProvenanceURL:     "census.gov",
					},
					{
						Val:               map[string]float64{"2011": 101, "2012": 102, "2013": 103},
						MeasurementMethod: "CensusACS5yrSurvey",
						ImportName:        "CensusACS5YearSurvey",
						ProvenanceURL:     "census.gov",
					},
				},
			},
			"",
			"",
			"",
			&ObsTimeSeries{
				Data:          map[string]float64{"2011": 100, "2012": 101},
				ProvenanceURL: "census.gov",
			},
		},
		// Filter by mmethod
		{
			&ObsTimeSeries{
				SourceSeries: []*SourceSeries{
					{
						Val:               map[string]float64{"2011": 100, "2012": 101},
						MeasurementMethod: "CensusPEPSurvey",
						ImportName:        "CensusPEP",
						ProvenanceURL:     "census.gov",
					},
					{
						Val:               map[string]float64{"2011": 101, "2012": 102, "2013": 103},
						MeasurementMethod: "CensusACS5yrSurvey",
						ImportName:        "CensusACS5YearSurvey",
						ProvenanceURL:     "census.gov",
					},
				},
			},
			"CensusACS5yrSurvey",
			"",
			"",
			&ObsTimeSeries{
				Data:          map[string]float64{"2011": 101, "2012": 102, "2013": 103},
				ProvenanceURL: "census.gov",
			},
		},
		// Filter by observation period
		{
			&ObsTimeSeries{
				SourceSeries: []*SourceSeries{
					{
						Val:               map[string]float64{"2011": 100, "2012": 101},
						MeasurementMethod: "CensusPEPSurvey",
						ImportName:        "CensusPEP",
						ObservationPeriod: "P1Y",
						ProvenanceURL:     "census.gov",
					},
					{
						Val:               map[string]float64{"2017": 101},
						MeasurementMethod: "CensusPEPSurvey",
						ImportName:        "CensusPEP",
						ObservationPeriod: "P2Y",
						ProvenanceURL:     "census.gov",
					},
				},
			},
			"",
			"",
			"P2Y",
			&ObsTimeSeries{
				Data:          map[string]float64{"2017": 101},
				ProvenanceURL: "census.gov",
			},
		},
		// No match
		{
			&ObsTimeSeries{
				SourceSeries: []*SourceSeries{
					{
						Val:               map[string]float64{"2011": 100, "2012": 101},
						MeasurementMethod: "CensusPEPSurvey",
						ImportName:        "CensusPEP",
						ObservationPeriod: "P1Y",
						ProvenanceURL:     "census.gov",
					},
					{
						Val:               map[string]float64{"2017": 101},
						MeasurementMethod: "CensusPEPSurvey",
						ImportName:        "CensusPEP",
						ObservationPeriod: "P2Y",
						ProvenanceURL:     "census.gov",
					},
				},
			},
			"",
			"",
			"P3Y",
			&ObsTimeSeries{},
		},
	} {
		got := c.input
		got.filterAndRank(&ObsProp{
			Mmethod: c.mmethod,
			Operiod: c.op,
			Unit:    c.unit})
		if diff := cmp.Diff(got, c.want, protocmp.Transform()); diff != "" {
			t.Errorf("filterAndRank() got diff %+v", diff)
		}
	}
}

func TestByRank(t *testing.T) {
	sourceSeries := []*SourceSeries{
		{
			Val:               map[string]float64{"2011": 101, "2012": 102, "2013": 103},
			MeasurementMethod: "randomeMMethod",
			ImportName:        "randomImportName",
			ProvenanceURL:     "census.gov",
		},
		{
			Val:               map[string]float64{"2011": 101, "2012": 102, "2013": 103},
			MeasurementMethod: "CensusACS5yrSurvey",
			ImportName:        "CensusACS5YearSurvey",
			ProvenanceURL:     "census.gov",
		},
		{
			Val:               map[string]float64{"2011": 100, "2012": 101},
			MeasurementMethod: "CensusPEPSurvey",
			ImportName:        "CensusPEP",
			ProvenanceURL:     "census.gov",
		},
	}
	sort.Sort(byRank(sourceSeries))
	expectImportName := []string{"CensusPEP", "CensusACS5YearSurvey", "randomImportName"}
	for index, series := range sourceSeries {
		if expectImportName[index] != series.ImportName {
			t.Errorf("Bad ranking for %d, %s", index, series.ImportName)
		}
	}
}

func TestGetLatest(t *testing.T) {
	obsTimeSeries := &ObsTimeSeries{
		SourceSeries: []*SourceSeries{
			{
				Val:               map[string]float64{"2011": 101, "2012": 102, "2013": 105, "2014": 200},
				MeasurementMethod: "randomeMMethod",
				ImportName:        "randomImportName",
				ProvenanceURL:     "census.gov",
			},
			{
				Val:               map[string]float64{"2011": 101, "2012": 102, "2013": 103},
				MeasurementMethod: "CensusACS5yrSurvey",
				ImportName:        "CensusACS5YearSurvey",
				ProvenanceURL:     "census.gov",
			},
			{
				Val:               map[string]float64{"2011": 100, "2012": 101},
				MeasurementMethod: "CensusPEPSurvey",
				ImportName:        "CensusPEP",
				ProvenanceURL:     "census.gov",
			},
		},
	}

	for _, c := range []struct {
		date string
		want float64
	}{
		{
			"",
			200,
		},
		{
			"2013",
			103,
		},
		{
			"2014",
			200,
		},
	} {
		value, _ := getValueFromBestSource(obsTimeSeries, c.date)
		if c.want != value {
			t.Errorf("Wrong latest value %f", value)
		}
	}
}
