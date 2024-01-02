package ofctrl

import (
	"fmt"
	"net"

	"antrea.io/libOpenflow/openflow15"
	"antrea.io/libOpenflow/util"
)

type GroupType int

const (
	GroupAll GroupType = iota
	GroupSelect
	GroupIndirect
	GroupFF
)

type GroupBundleMessage struct {
	message *openflow15.GroupMod
}

func (m *GroupBundleMessage) resetXid(xid uint32) util.Message {
	m.message.Xid = xid
	return m.message
}

func (m *GroupBundleMessage) getXid() uint32 {
	return m.message.Xid
}

func (m *GroupBundleMessage) GetMessage() util.Message {
	return m.message
}

type Group struct {
	Switch      *OFSwitch
	ID          uint32
	GroupType   GroupType
	Buckets     []*openflow15.Bucket
	Properties  []util.Message
	isInstalled bool
}

func (self *Group) Type() string {
	return "group"
}

func (self *Group) GetActionMessage() openflow15.Action {
	return openflow15.NewActionGroup(self.ID)
}

func (self *Group) GetActionType() string {
	return ActTypeGroup
}

func (self *Group) GetFlowInstr() openflow15.Instruction {
	groupInstr := openflow15.NewInstrApplyActions()
	groupAct := self.GetActionMessage()
	// Add group action to the instruction
	groupInstr.AddAction(groupAct, false)
	return groupInstr
}

func (self *Group) AddBuckets(buckets ...*openflow15.Bucket) {
	if self.Buckets == nil {
		self.Buckets = make([]*openflow15.Bucket, 0)
	}
	self.Buckets = append(self.Buckets, buckets...)
	if self.isInstalled {
		self.Install()
	}
}

func (self *Group) ResetBuckets(buckets ...*openflow15.Bucket) {
	self.Buckets = make([]*openflow15.Bucket, 0)
	self.Buckets = append(self.Buckets, buckets...)
	if self.isInstalled {
		self.Install()
	}
}

func (self *Group) AddProperty(prop util.Message) {
	self.Properties = append(self.Properties, prop)
	if self.isInstalled {
		self.Install()
	}
}

func (self *Group) Install() error {
	command := openflow15.OFPGC_ADD
	if self.isInstalled {
		command = openflow15.OFPGC_MODIFY
	}
	groupMod := self.getGroupModMessage(command)

	if err := self.Switch.Send(groupMod); err != nil {
		return err
	}

	// Mark it as installed
	self.isInstalled = true

	return nil
}

func (self *Group) getGroupModMessage(command int) *openflow15.GroupMod {
	groupMod := openflow15.NewGroupMod()
	groupMod.GroupId = self.ID
	groupMod.Command = uint16(command)

	switch self.GroupType {
	case GroupAll:
		groupMod.Type = openflow15.GT_ALL
	case GroupSelect:
		groupMod.Type = openflow15.GT_SELECT
	case GroupIndirect:
		groupMod.Type = openflow15.GT_INDIRECT
	case GroupFF:
		groupMod.Type = openflow15.GT_FF
	}

	if command == openflow15.OFPGC_DELETE {
		return groupMod
	}

	if command == openflow15.OFPGC_ADD || command == openflow15.OFPGC_MODIFY {
		for _, prop := range self.Properties {
			groupMod.Properties = append(groupMod.Properties, prop)
		}
	}

	for _, bkt := range self.Buckets {
		// Add the bucket to group
		groupMod.AddBucket(*bkt)
	}

	if command == openflow15.OFPGC_INSERT_BUCKET {
		groupMod.CommandBucketId = openflow15.OFPG_BUCKET_LAST
	}

	return groupMod
}

func (self *Group) GetBundleMessage(command int) *GroupBundleMessage {
	groupMod := self.getGroupModMessage(command)
	return &GroupBundleMessage{groupMod}
}

func (self *Group) Delete() error {
	if self.isInstalled {
		groupMod := openflow15.NewGroupMod()
		groupMod.GroupId = self.ID
		groupMod.Command = openflow15.OFPGC_DELETE
		if err := self.Switch.Send(groupMod); err != nil {
			return err
		}
		// Mark it as unInstalled
		self.isInstalled = false
	}

	// Delete group from switch cache
	return self.Switch.DeleteGroup(self.ID)
}

func NewGroup(groupId uint32, groupType GroupType, sw *OFSwitch) *Group {
	return &Group{
		ID:        groupId,
		GroupType: groupType,
		Switch:    sw,
	}
}

const (
	GroupHashSrcIp = iota
	GroupHashDstIp
	GroupHashSrcPort
	GroupHashDstPort
	GroupHashProtocol
)

func (self *Group) SetSelectionMethod(hashName string, param uint64, fields ...int) {
	mt := openflow15.NTRSelectionMethodType(hashName)

	matches := []openflow15.MatchField{}
	for _, fd := range fields {
		switch fd {
		case GroupHashSrcIp:
			srcIp := openflow15.NewIpv4SrcField(net.IPv4bcast, nil)
			matches = append(matches, *srcIp)
		case GroupHashDstIp:
			dstIp := openflow15.NewIpv4DstField(net.IPv4bcast, nil)
			matches = append(matches, *dstIp)
		case GroupHashSrcPort:
			sp := openflow15.NewTcpSrcField(0xffff)
			matches = append(matches, *sp)
		case GroupHashDstPort:
			dp := openflow15.NewTcpDstField(0xffff)
			matches = append(matches, *dp)
		case GroupHashProtocol:
			prot := openflow15.NewIpProtoField(0xff)
			matches = append(matches, *prot)
		}
	}

	property := openflow15.NewNTRSelectionMethod(mt, param, matches...)
	self.AddProperty(property)
}

func (self *Group) AddBucket(bkt *Bucket) error {
	// convert the Flow actions to the Bucket
	command := openflow15.FC_ADD
	bkt.Flow.Table = self.Switch.DefaultTable()
	flowMod, err := bkt.GenerateFlowModMessage(command)
	if err != nil {
		return err
	}

	for _, inst := range flowMod.Instructions {
		acts, ok := inst.(*openflow15.InstrActions)
		if !ok || acts == nil {
			return fmt.Errorf("wrong action type for Bucket: %T", inst)
		}

		for _, act := range acts.Actions {
			bkt.AddAction(act)
		}
	}

	self.AddBuckets(&bkt.Bucket)
	return nil
}
