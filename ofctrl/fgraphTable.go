/*
Copyright 2014 Cisco Systems Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package ofctrl

// This file implements the forwarding graph API for the table

import (
	"fmt"
	"sync"

	"antrea.io/libOpenflow/openflow15"
)

// Fgraph table element
type Table struct {
	Switch  *OFSwitch
	TableId uint8
	flowDb  map[string]*Flow // database of flow entries
	lock    sync.Mutex       // lock flowdb modification
}

// Fgraph element type for table
func (self *Table) Type() string {
	return "table"
}

// instruction set for table element
func (self *Table) GetFlowInstr() openflow15.Instruction {
	return openflow15.NewInstrGotoTable(self.TableId)
}

// FIXME: global unique flow cookie
var globalFlowID uint64 = 1

// Create a new flow on the table
func (self *Table) NewFlow(match FlowMatch) (*Flow, error) {
	// modifications to flowdb requires a lock
	self.lock.Lock()
	defer self.lock.Unlock()
	logger := self.Switch.logger

	flow := new(Flow)
	flow.Table = self
	flow.Match = match
	flow.isInstalled = false
	flow.flowActions = make([]*FlowAction, 0)

	logger.Debugf("Creating new flow for match: %+v", match)

	// See if the flow already exists
	flowKey := flow.flowKey()
	if self.flowDb[flowKey] != nil {
		logger.Errorf("Flow %s already exists", flowKey)
		return nil, fmt.Errorf("Flow %s already exists", flowKey)
	}

	logger.Debugf("Added flow: %s", flowKey)

	// Save it in DB. We dont install the flow till its next graph elem is set
	self.flowDb[flowKey] = flow

	return flow, nil
}

// Delete a flow from the table
func (self *Table) DeleteFlow(flowKey string) error {
	// modifications to flowdb requires a lock
	self.lock.Lock()
	defer self.lock.Unlock()
	logger := self.Switch.logger

	// first empty it and then delete it.
	self.flowDb[flowKey] = nil
	delete(self.flowDb, flowKey)

	logger.Debugf("Deleted flow: %s", flowKey)

	return nil
}

// Delete the table
func (self *Table) Delete() error {
	// FIXME: Delete the table
	return nil
}

func NewTable(tableId uint8, sw *OFSwitch) *Table {
	table := new(Table)
	table.Switch = sw
	table.TableId = tableId
	return table
}
