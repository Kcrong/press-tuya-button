package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tuya/tuya-cloud-sdk-go/api/common"
	"github.com/tuya/tuya-cloud-sdk-go/api/device"
	"github.com/tuya/tuya-cloud-sdk-go/config"
)

const (
	deviceID  = "removed"
	accessID  = "removed"
	accessKey = "removed"
	sentryDSN = "removed"

	defaultPort = "8012"
)

func main() {
	config.SetEnv(
		common.URLUS,
		accessID,
		accessKey,
	)

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:           sentryDSN,
		EnableTracing: true,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %+v", errors.WithStack(err))
	}

	r := gin.Default()
	r.Use(sentrygin.New(sentrygin.Options{}))

	r.GET("/press", func(c *gin.Context) {
		res, err := device.PostDeviceCommand(deviceID, []device.Command{
			{
				Code:  "switch",
				Value: true,
			},
		})

		if err == nil && res.Success {
			c.Status(http.StatusNoContent)
			return
		}

		if err != nil {
			sentry.CaptureException(errors.WithStack(err))
		}
		if !res.Success {
			sentry.CaptureMessage("failed to press: " + fmt.Sprintf("%+v", res))
		}

		c.Status(http.StatusInternalServerError)
	})

	port := os.Getenv("GIN_PORT")
	if port == "" {
		port = defaultPort
	}

	if err := r.Run(":" + port); err != nil {
		logrus.Fatalf("%+v", err)
	}
}
