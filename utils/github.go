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
	signature, err := hex.DecodeString(strings.Split(req.Headers["X-Hub-Signature"], "=")[1])
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
