package chat

// MessagePusher abstracts downstream message delivery to connected clients,
// enabling transparent switching between Monolith Hub (in-memory WebSocket dispatch)
// and Microservices ClusterPusher (gRPC / Kafka push to distributed Gateway instances).
//
// Return value semantics of SendToUser:
// - Monolith Hub: returns true if the frame was successfully buffered into local client connection channel(s).
// - ClusterPusher: returns true if the frame was successfully handed off into the downstream delivery pipeline
//   (published to Kafka TopicPush or accepted by target Gateway gRPC). It indicates pipeline acceptance,
//   NOT client-side delivery confirmation.
//
// End-to-end client delivery confirmation is strictly tracked by the ACK state machine:
// CmdDeliveryAck (0x0103) -> MsgStatusDelivered (2) and CmdReadReceipt (0x0401) -> MsgStatusRead (3).
type MessagePusher interface {
	SendToUser(userID uint64, data []byte) bool
	SendToUserExclude(userID uint64, excludeDeviceID string, data []byte)
}
