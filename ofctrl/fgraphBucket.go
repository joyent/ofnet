package ofctrl

import (
	"net"

	"antrea.io/libOpenflow/openflow15"
)

type Bucket struct {
	openflow15.Bucket
	Flow
}

// Create a new Bucket
func NewBucket(bktId uint32) *Bucket {
	return &Bucket{
		Bucket: openflow15.Bucket{
			BucketId: bktId,
		},
	}
}

func (self *Bucket) SetOutput(portNum uint32) {
	// Create a new output element
	output := new(Output)
	output.outputType = "port"
	output.portNo = portNum

	self.NextElem = output
}

func (self *Bucket) SetTunnelSrcIp(srcIp string) {
	// Set tunnel src addr field
	ip := net.ParseIP(srcIp)
	if ip.IsUnspecified() || ip.IsLoopback() {
		return
	}

	self.SetIPField(ip, "TunSrc")
}

func (self *Bucket) SetTunnelDstIp(dstIp string) {
	// Set tunnel dst addr field
	ip := net.ParseIP(dstIp)
	if ip.IsUnspecified() || ip.IsLoopback() {
		return
	}

	self.SetIPField(ip, "TunDst")
}

func (self *Bucket) SetWeight(weight uint16) {
	wt := openflow15.NewGroupBucketPropWeight(weight)
	self.AddProperty(wt)
}

func (self *Bucket) SetWatchPort(portNum uint32) {
	wt := openflow15.NewGroupBucketPropWatchPort(portNum)
	self.AddProperty(wt)
}

func (self *Bucket) SetWatchGroup(groupId uint32) {
	wt := openflow15.NewGroupBucketPropWatchGroup(groupId)
	self.AddProperty(wt)
}
