package firebase

import (
	"context"
	"fmt"
	"os"
	"osush/internal/handlers/request"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

const credentialsFile = "internal/credentials/osushecommerce-firebase-adminsdk-fbsvc-c8f00caf54.json"

type FirebaseService struct {
	client *messaging.Client
}

// credentialsOption prefers the FIREBASE_CREDENTIALS_JSON environment variable
// and falls back to the local, git-ignored credentials file.
func credentialsOption() option.ClientOption {
	if credentialsJSON := os.Getenv("FIREBASE_CREDENTIALS_JSON"); credentialsJSON != "" {
		return option.WithCredentialsJSON([]byte(credentialsJSON))
	}

	return option.WithCredentialsFile(credentialsFile)
}

func NewFirebaseService(ctx context.Context) (*FirebaseService, error) {
	app, err := firebase.NewApp(
		ctx,
		nil,
		credentialsOption(),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to initialize firebase app: %w",
			err,
		)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to initialize firebase messaging: %w",
			err,
		)
	}

	return &FirebaseService{
		client: client,
	}, nil
}

func (f *FirebaseService) SendNotifications(
	ctx context.Context,
	deviceTokens []string,
	title request.Dictionary,
	body request.Dictionary,
) error {

	if len(deviceTokens) == 0 {
		return nil
	}

	const batchSize = 500

	for i := 0; i < len(deviceTokens); i += batchSize {

		end := i + batchSize

		if end > len(deviceTokens) {
			end = len(deviceTokens)
		}

		tokens := deviceTokens[i:end]

		message := &messaging.MulticastMessage{
			Tokens: tokens,
			Notification: &messaging.Notification{
				Title: title.Tm,
				Body:  body.Tm,
			},
			Data: map[string]string{
				"title_tm": title.Tm,
				"title_ru": title.Ru,
				"body_tm":  body.Tm,
				"body_ru":  body.Ru,
			},
		}

		response, err := f.client.SendEachForMulticast(
			ctx,
			message,
		)
		if err != nil {
			return fmt.Errorf(
				"failed to send notifications: %w",
				err,
			)
		}

		fmt.Printf(
			"firebase notifications: success=%d failure=%d\n",
			response.SuccessCount,
			response.FailureCount,
		)
	}

	return nil
}

// SendToTokens sends one notification with custom data (e.g. route and
// order_id so the app opens the right page on tap) to the given tokens.
// Used for per-client pushes such as "your order has been sent".
func (f *FirebaseService) SendToTokens(
	ctx context.Context,
	deviceTokens []string,
	title string,
	body string,
	data map[string]string,
) (*messaging.BatchResponse, error) {
	if len(deviceTokens) == 0 {
		return &messaging.BatchResponse{}, nil
	}

	message := &messaging.MulticastMessage{
		Tokens: deviceTokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				// Must match the channel created by the Flutter app.
				ChannelID: "high_importance_channel",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{Sound: "default"},
			},
		},
	}

	response, err := f.client.SendEachForMulticast(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send notifications: %w", err)
	}

	return response, nil
}
