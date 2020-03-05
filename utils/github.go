package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// VerifyGithubWebhookRequest validate X-Hub-Signature
func VerifyGithubWebhookRequest(req events.APIGatewayProxyRequest, secret string) error {
	xHubSignature, ok := req.Headers["X-Hub-Signature"]
	if !ok {
		return fmt.Errorf("Missing signature")
	}
	signatureParts := strings.Split(xHubSignature, "=")
	if len(signatureParts) < 2 {
		return fmt.Errorf("Bad signature format")
	}
	signature, err := hex.DecodeString(signatureParts[1])
	if err != nil {
		return err
	}
	if !verifyPayload(secret, []byte(req.Body), signature) {
		return fmt.Errorf("Signature missmatch")
	}

	return nil
}

func verifyPayload(secret string, payload, signature []byte) bool {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(signature, expectedMAC)
}
