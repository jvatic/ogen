package integration

import (
	"fmt"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_discriminator_mapping"
)

func TestDiscriminatorMapping(t *testing.T) {
	t.Run("EventSum", func(t *testing.T) {
		t.Run("MultipleDiscriminatorValues", func(t *testing.T) {
			// Test that multiple discriminator values map to the same type
			for i, tc := range []struct {
				Input         string
				ExpectedType  api.EventSumType
				IsUserEvent   bool
				IsAdminEvent  bool
				IsSystemEvent bool
			}{
				{
					`{"type": "user.created", "timestamp": "2023-01-01T00:00:00Z", "userId": "user123"}`,
					api.EventSumUserCreated,
					true, false, false,
				},
				{
					`{"type": "user.registered", "timestamp": "2023-01-01T00:00:00Z", "userId": "user123"}`,
					api.EventSumUserRegistered,
					true, false, false,
				},
				{
					`{"type": "user.signup", "timestamp": "2023-01-01T00:00:00Z", "userId": "user123"}`,
					api.EventSumUserSignup,
					true, false, false,
				},
				{
					`{"type": "admin.login", "timestamp": "2023-01-01T00:00:00Z", "adminId": "admin123"}`,
					api.EventSumAdminLogin,
					false, true, false,
				},
				{
					`{"type": "admin.action", "timestamp": "2023-01-01T00:00:00Z", "adminId": "admin123"}`,
					api.EventSumAdminAction,
					false, true, false,
				},
				{
					`{"type": "system.startup", "timestamp": "2023-01-01T00:00:00Z", "service": "api"}`,
					api.EventSumSystemStartup,
					false, false, true,
				},
				{
					`{"type": "system.shutdown", "timestamp": "2023-01-01T00:00:00Z", "service": "api"}`,
					api.EventSumSystemShutdown,
					false, false, true,
				},
				{
					`{"type": "system.maintenance", "timestamp": "2023-01-01T00:00:00Z", "service": "api"}`,
					api.EventSumSystemMaintenance,
					false, false, true,
				},
			} {
				tc := tc
				t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
					var event api.Event
					require.NoError(t, event.Decode(jx.DecodeStr(tc.Input)))

					// Check discriminator type
					require.Equal(t, tc.ExpectedType, event.OneOf.Type)

					// Check Is*() methods
					require.Equal(t, tc.IsUserEvent, event.OneOf.IsUserEvent())
					require.Equal(t, tc.IsAdminEvent, event.OneOf.IsAdminEvent())
					require.Equal(t, tc.IsSystemEvent, event.OneOf.IsSystemEvent())

					// Test round-trip encoding
					testEncodeEvent(t, &event, tc.Input)
				})
			}
		})

		t.Run("SetWithType", func(t *testing.T) {
			// Test the new SetWithType methods
			var eventSum api.EventSum

			// Test UserEvent with different discriminator values
			userEvent := api.UserEvent{
				UserId: "user123",
				Email:  api.NewOptString("user@example.com"),
			}

			eventSum.SetUserEventWithType(userEvent, api.EventSumUserCreated)
			require.Equal(t, api.EventSumUserCreated, eventSum.Type)
			require.True(t, eventSum.IsUserEvent())

			eventSum.SetUserEventWithType(userEvent, api.EventSumUserRegistered)
			require.Equal(t, api.EventSumUserRegistered, eventSum.Type)
			require.True(t, eventSum.IsUserEvent())

			// Test AdminEvent with different discriminator values
			adminEvent := api.AdminEvent{
				AdminId: "admin123",
				Action:  api.NewOptString("login"),
			}

			eventSum.SetAdminEventWithType(adminEvent, api.EventSumAdminLogin)
			require.Equal(t, api.EventSumAdminLogin, eventSum.Type)
			require.True(t, eventSum.IsAdminEvent())

			eventSum.SetAdminEventWithType(adminEvent, api.EventSumAdminAction)
			require.Equal(t, api.EventSumAdminAction, eventSum.Type)
			require.True(t, eventSum.IsAdminEvent())
		})

		t.Run("DefaultSetMethods", func(t *testing.T) {
			// Test that default Set methods pick the first discriminator value
			var eventSum api.EventSum

			userEvent := api.UserEvent{UserId: "user123"}
			eventSum.SetUserEvent(userEvent)
			require.Equal(t, api.EventSumUserCreated, eventSum.Type)

			adminEvent := api.AdminEvent{AdminId: "admin123"}
			eventSum.SetAdminEvent(adminEvent)
			require.Equal(t, api.EventSumAdminAction, eventSum.Type)

			systemEvent := api.SystemEvent{Service: "api"}
			eventSum.SetSystemEvent(systemEvent)
			require.Equal(t, api.EventSumSystemMaintenance, eventSum.Type)
		})
	})

	t.Run("NotificationSum", func(t *testing.T) {
		t.Run("MultipleDiscriminatorValues", func(t *testing.T) {
			for i, tc := range []struct {
				Input               string
				ExpectedType        api.NotificationSumType
				IsEmailNotification bool
				IsSmsNotification   bool
				IsPushNotification  bool
			}{
				{
					`{"channel": "email", "message": "Hello", "recipient": "user@example.com"}`,
					api.NotificationSumEmail,
					true, false, false,
				},
				{
					`{"channel": "mail", "message": "Hello", "recipient": "user@example.com"}`,
					api.NotificationSumMail,
					true, false, false,
				},
				{
					`{"channel": "sms", "message": "Hello", "phoneNumber": "+1234567890"}`,
					api.NotificationSumSMS,
					false, true, false,
				},
				{
					`{"channel": "text", "message": "Hello", "phoneNumber": "+1234567890"}`,
					api.NotificationSumText,
					false, true, false,
				},
				{
					`{"channel": "push", "message": "Hello", "deviceToken": "token123"}`,
					api.NotificationSumPush,
					false, false, true,
				},
				{
					`{"channel": "mobile", "message": "Hello", "deviceToken": "token123"}`,
					api.NotificationSumMobile,
					false, false, true,
				},
				{
					`{"channel": "app", "message": "Hello", "deviceToken": "token123"}`,
					api.NotificationSumApp,
					false, false, true,
				},
			} {
				tc := tc
				t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
					var notification api.Notification
					require.NoError(t, notification.Decode(jx.DecodeStr(tc.Input)))

					// Check discriminator type
					require.Equal(t, tc.ExpectedType, notification.OneOf.Type)

					// Check Is*() methods
					require.Equal(t, tc.IsEmailNotification, notification.OneOf.IsEmailNotification())
					require.Equal(t, tc.IsSmsNotification, notification.OneOf.IsSmsNotification())
					require.Equal(t, tc.IsPushNotification, notification.OneOf.IsPushNotification())

					// Test round-trip encoding
					testEncodeEvent(t, &notification, tc.Input)
				})
			}
		})

		t.Run("SetWithType", func(t *testing.T) {
			var notificationSum api.NotificationSum

			// Test EmailNotification with different discriminator values
			emailNotification := api.EmailNotification{
				Recipient: "user@example.com",
				Subject:   api.NewOptString("Test Subject"),
			}

			notificationSum.SetEmailNotificationWithType(emailNotification, api.NotificationSumEmail)
			require.Equal(t, api.NotificationSumEmail, notificationSum.Type)
			require.True(t, notificationSum.IsEmailNotification())

			notificationSum.SetEmailNotificationWithType(emailNotification, api.NotificationSumMail)
			require.Equal(t, api.NotificationSumMail, notificationSum.Type)
			require.True(t, notificationSum.IsEmailNotification())

			// Test SmsNotification with different discriminator values
			smsNotification := api.SmsNotification{
				PhoneNumber: "+1234567890",
			}

			notificationSum.SetSmsNotificationWithType(smsNotification, api.NotificationSumSMS)
			require.Equal(t, api.NotificationSumSMS, notificationSum.Type)
			require.True(t, notificationSum.IsSmsNotification())

			notificationSum.SetSmsNotificationWithType(smsNotification, api.NotificationSumText)
			require.Equal(t, api.NotificationSumText, notificationSum.Type)
			require.True(t, notificationSum.IsSmsNotification())
		})
	})

	t.Run("Validation", func(t *testing.T) {
		// Test validation of discriminator mappings
		for i, tc := range []struct {
			Input   string
			Valid   bool
			IsEvent bool
		}{
			{
				`{"type": "user.created", "timestamp": "2023-01-01T00:00:00Z", "userId": "user123"}`,
				true,
				true,
			},
			{
				`{"type": "invalid.type", "timestamp": "2023-01-01T00:00:00Z", "userId": "user123"}`,
				false,
				true,
			},
			{
				`{"channel": "email", "message": "Hello", "recipient": "user@example.com"}`,
				true,
				false,
			},
			{
				`{"channel": "invalid.channel", "message": "Hello", "recipient": "user@example.com"}`,
				false,
				false,
			},
		} {
			tc := tc
			t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
				if tc.IsEvent {
					var event api.Event
					err := event.Decode(jx.DecodeStr(tc.Input))
					if tc.Valid {
						require.NoError(t, err)
					} else {
						require.Error(t, err)
					}
				} else {
					var notification api.Notification
					err := notification.Decode(jx.DecodeStr(tc.Input))
					if tc.Valid {
						require.NoError(t, err)
					} else {
						require.Error(t, err)
					}
				}
			})
		}
	})

	t.Run("AlertSum_AnyOf", func(t *testing.T) {
		t.Run("MultipleDiscriminatorValues", func(t *testing.T) {
			// Test anyOf with discriminator mappings
			for i, tc := range []struct {
				Input                string
				ExpectedType         api.AlertSumType
				IsSecurityAlert      bool
				IsPerformanceAlert   bool
				IsInfoAlert          bool
			}{
				{
					`{"severity": "critical", "message": "Security breach", "timestamp": "2023-01-01T00:00:00Z", "threatLevel": "critical", "source": "firewall"}`,
					api.AlertSumCritical,
					true, false, false,
				},
				{
					`{"severity": "high", "message": "Security issue", "timestamp": "2023-01-01T00:00:00Z", "threatLevel": "high", "source": "ids"}`,
					api.AlertSumHigh,
					true, false, false,
				},
				{
					`{"severity": "urgent", "message": "Urgent security", "timestamp": "2023-01-01T00:00:00Z", "threatLevel": "high", "source": "scanner"}`,
					api.AlertSumUrgent,
					true, false, false,
				},
				{
					`{"severity": "medium", "message": "Performance issue", "timestamp": "2023-01-01T00:00:00Z", "metric": "cpu", "threshold": 80.0}`,
					api.AlertSumMedium,
					false, true, false,
				},
				{
					`{"severity": "low", "message": "Low performance", "timestamp": "2023-01-01T00:00:00Z", "metric": "memory", "threshold": 90.0}`,
					api.AlertSumLow,
					false, true, false,
				},
				{
					`{"severity": "info", "message": "Info message", "timestamp": "2023-01-01T00:00:00Z", "category": "system"}`,
					api.AlertSumInfo,
					false, false, true,
				},
				{
					`{"severity": "debug", "message": "Debug info", "timestamp": "2023-01-01T00:00:00Z", "category": "debug"}`,
					api.AlertSumDebug,
					false, false, true,
				},
			} {
				tc := tc
				t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
					var alert api.Alert
					require.NoError(t, alert.Decode(jx.DecodeStr(tc.Input)))
					
					// Check discriminator type
					require.Equal(t, tc.ExpectedType, alert.AnyOf.Type)
					
					// Check Is*() methods
					require.Equal(t, tc.IsSecurityAlert, alert.AnyOf.IsSecurityAlert())
					require.Equal(t, tc.IsPerformanceAlert, alert.AnyOf.IsPerformanceAlert())
					require.Equal(t, tc.IsInfoAlert, alert.AnyOf.IsInfoAlert())
					
					// Test round-trip encoding
					testEncodeEvent(t, &alert, tc.Input)
				})
			}
		})

		t.Run("SetWithType", func(t *testing.T) {
			var alertSum api.AlertSum
			
			// Test SecurityAlert with different discriminator values
			securityAlert := api.SecurityAlert{
				ThreatLevel: api.SecurityAlertThreatLevelCritical,
				Source:      "test-source",
			}
			
			alertSum.SetSecurityAlertWithType(securityAlert, api.AlertSumCritical)
			require.Equal(t, api.AlertSumCritical, alertSum.Type)
			require.True(t, alertSum.IsSecurityAlert())
			
			alertSum.SetSecurityAlertWithType(securityAlert, api.AlertSumHigh)
			require.Equal(t, api.AlertSumHigh, alertSum.Type)
			require.True(t, alertSum.IsSecurityAlert())
			
			alertSum.SetSecurityAlertWithType(securityAlert, api.AlertSumUrgent)
			require.Equal(t, api.AlertSumUrgent, alertSum.Type)
			require.True(t, alertSum.IsSecurityAlert())
			
			// Test PerformanceAlert with different discriminator values
			performanceAlert := api.PerformanceAlert{
				Metric:    "cpu",
				Threshold: 75.0,
			}
			
			alertSum.SetPerformanceAlertWithType(performanceAlert, api.AlertSumMedium)
			require.Equal(t, api.AlertSumMedium, alertSum.Type)
			require.True(t, alertSum.IsPerformanceAlert())
			
			alertSum.SetPerformanceAlertWithType(performanceAlert, api.AlertSumLow)
			require.Equal(t, api.AlertSumLow, alertSum.Type)
			require.True(t, alertSum.IsPerformanceAlert())
		})

		t.Run("DefaultSetMethods", func(t *testing.T) {
			// Test that default Set methods pick the first discriminator value for anyOf
			var alertSum api.AlertSum
			
			securityAlert := api.SecurityAlert{
				ThreatLevel: api.SecurityAlertThreatLevelHigh,
				Source:      "test",
			}
			alertSum.SetSecurityAlert(securityAlert)
			require.Equal(t, api.AlertSumCritical, alertSum.Type) // First value alphabetically
			
			performanceAlert := api.PerformanceAlert{
				Metric:    "memory",
				Threshold: 80.0,
			}
			alertSum.SetPerformanceAlert(performanceAlert)
			require.Equal(t, api.AlertSumLow, alertSum.Type) // First value alphabetically
			
			infoAlert := api.InfoAlert{Category: "system"}
			alertSum.SetInfoAlert(infoAlert)
			require.Equal(t, api.AlertSumDebug, alertSum.Type) // First value alphabetically
		})
	})

	t.Run("GetMethods", func(t *testing.T) {
		// Test Get*() methods with discriminator mappings
		var eventSum api.EventSum
		
		userEvent := api.UserEvent{UserId: "user123"}
		eventSum.SetUserEvent(userEvent)
		
		gotUserEvent, ok := eventSum.GetUserEvent()
		require.True(t, ok)
		require.Equal(t, userEvent.UserId, gotUserEvent.UserId)
		
		_, ok = eventSum.GetAdminEvent()
		require.False(t, ok)
		
		_, ok = eventSum.GetSystemEvent()
		require.False(t, ok)
	})
}

func TestDiscriminatorMappingEncodeDecodeRoundTrip(t *testing.T) {
	// Test manual encode/decode round trips
	t.Run("EventSum", func(t *testing.T) {
		userEvent := api.UserEvent{UserId: "user123"}
		eventSum := api.NewUserEventEventSum(userEvent)
		testEncodeDecode(t, &eventSum)
	})
	
	t.Run("NotificationSum", func(t *testing.T) {
		emailNotification := api.EmailNotification{Recipient: "user@example.com"}
		notificationSum := api.NewEmailNotificationNotificationSum(emailNotification)
		testEncodeDecode(t, &notificationSum)
	})

	t.Run("AlertSum_AnyOf", func(t *testing.T) {
		securityAlert := api.SecurityAlert{
			ThreatLevel: api.SecurityAlertThreatLevelHigh,
			Source:      "test-source",
		}
		alertSum := api.NewSecurityAlertAlertSum(securityAlert)
		testEncodeDecode(t, &alertSum)
	})
}

func testEncodeDecode(t *testing.T, v interface {
	Encode(*jx.Encoder)
	Decode(*jx.Decoder) error
}) {
	e := jx.Encoder{}
	v.Encode(&e)
	data := e.Bytes()

	d := jx.DecodeBytes(data)
	require.NoError(t, v.Decode(d))
}

func testEncodeEvent(t *testing.T, v interface {
	Encode(*jx.Encoder)
}, expectedJSON string) {
	e := jx.Encoder{}
	v.Encode(&e)
	data := e.Bytes()
	
	// Compare JSON semantically
	require.JSONEq(t, expectedJSON, string(data))
}
