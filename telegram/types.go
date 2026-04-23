package telegram

// SendVerificationMessageRequest is the main request payload for sendVerificationMessage.
// SendVerificationMessageRequest là payload chính cho sendVerificationMessage.
type SendVerificationMessageRequest = GatewaySendVerificationMessageRequest

// RequestStatus is the main response object returned by Telegram Gateway methods.
// RequestStatus là response object chính được Telegram Gateway trả về.
type RequestStatus = GatewayRequestStatus

// DeliveryStatus is the delivery status returned by Telegram Gateway.
// DeliveryStatus là trạng thái giao message do Telegram Gateway trả về.
type DeliveryStatus = GatewayDeliveryStatus

// VerificationStatus is the code verification status returned by Telegram Gateway.
// VerificationStatus là trạng thái xác thực mã do Telegram Gateway trả về.
type VerificationStatus = GatewayVerificationStatus
