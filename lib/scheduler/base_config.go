/*
 *
 * k6 - a next-generation load testing tool
 * Copyright (C) 2019 Load Impact
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package scheduler

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/loadimpact/k6/lib/types"
	null "gopkg.in/guregu/null.v3"
)

var schedulerNameWhitelist = regexp.MustCompile(`^[0-9a-zA-Z_-]+$`) //nolint:gochecknoglobals
const schedulerNameErr = "the scheduler name should contain only numbers, latin letters, underscores, and dashes"

// BaseConfig contains the common config fields for all schedulers
type BaseConfig struct {
	Name         string             `json:"-"` // set via the JS object key
	Type         string             `json:"type"`
	StartTime    types.NullDuration `json:"startTime"`
	GracefulStop types.NullDuration `json:"gracefulStop"`
	Env          map[string]string  `json:"env"`
	Exec         null.String        `json:"exec"` // function name, externally validated

	//TODO: future extensions like tags, distribution, others?
}

// NewBaseConfig returns a default base config with the default values
func NewBaseConfig(name, configType string) BaseConfig {
	return BaseConfig{
		Name:         name,
		Type:         configType,
		GracefulStop: types.NewNullDuration(30*time.Second, false),
	}
}

// Validate checks some basic things like present name, type, and a positive start time
func (bc BaseConfig) Validate() (errors []error) {
	// Some just-in-case checks, since those things are likely checked in other places or
	// even assigned by us:
	if bc.Name == "" {
		errors = append(errors, fmt.Errorf("scheduler name shouldn't be empty"))
	}
	if !schedulerNameWhitelist.MatchString(bc.Name) {
		errors = append(errors, fmt.Errorf(schedulerNameErr))
	}
	if bc.Exec.Valid && bc.Exec.String == "" {
		errors = append(errors, fmt.Errorf("exec value cannot be empty"))
	}
	if bc.Type == "" {
		errors = append(errors, fmt.Errorf("missing or empty type field"))
	}
	// The actually reasonable checks:
	if bc.StartTime.Duration < 0 {
		errors = append(errors, fmt.Errorf("the startTime can't be negative"))
	}
	if bc.GracefulStop.Duration < 0 {
		errors = append(errors, fmt.Errorf("the gracefulStop timeout can't be negative"))
	}
	return errors
}

// GetName returns the name of the scheduler.
func (bc BaseConfig) GetName() string {
	return bc.Name
}

// GetType returns the scheduler's type as a string ID.
func (bc BaseConfig) GetType() string {
	return bc.Type
}

// GetStartTime returns the starting time, relative to the beginning of the
// actual test, that this scheduler is supposed to execute.
func (bc BaseConfig) GetStartTime() time.Duration {
	return time.Duration(bc.StartTime.Duration)
}

// GetGracefulStop returns how long k6 is supposed to wait for any still
// running iterations to finish executing at the end of the normal scheduler
// duration, before it actually kills them.
//
// Of course, that doesn't count when the user manually interrupts the test,
// then iterations are immediately stopped.
func (bc BaseConfig) GetGracefulStop() time.Duration {
	return time.Duration(bc.GracefulStop.Duration)
}

// GetEnv returns any specific environment key=value pairs that
// are configured for the scheduler.
func (bc BaseConfig) GetEnv() map[string]string {
	return bc.Env
}

// GetExec returns the configured custom exec value, if any.
func (bc BaseConfig) GetExec() null.String {
	return bc.Exec
}

// IsDistributable returns true since by default all schedulers could be run in
// a distributed manner.
func (bc BaseConfig) IsDistributable() bool {
	return true
}

// getBaseInfo is a helper method for the the "parent" String methods.
func (bc BaseConfig) getBaseInfo(facts ...string) string {
	if bc.Exec.Valid {
		facts = append(facts, fmt.Sprintf("exec: %s", bc.Exec.String))
	}
	if bc.StartTime.Duration > 0 {
		facts = append(facts, fmt.Sprintf("startTime: %s", bc.StartTime.Duration))
	}
	if bc.GracefulStop.Duration > 0 {
		facts = append(facts, fmt.Sprintf("gracefulStop: %s", bc.GracefulStop.Duration))
	}
	if len(facts) == 0 {
		return ""
	}
	return " (" + strings.Join(facts, ", ") + ")"
}
