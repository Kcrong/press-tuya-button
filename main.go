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

	r.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", htmlIndex)
	})
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "https://raw.githubusercontent.com/gin-gonic/examples/master/favicon/favicon.ico")
	})

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

var (
	htmlIndex = []byte(`<!DOCTYPE html>

<html>
<head>
    <meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1"> 
    <title>배송 안내</title>
	<style>
		body { 
		  width: 100%; 
		  height: 100%; 
		} 
		html { 
		  width: 100%; 
		  height: 100%; 
		} 
		h1 {
		  font-weight: bold;
		  font-size: 80px;
		}
		p {
		  font-size: 40px;
		}
	</style>
</head>

<body>

<h1>배송 안내</h1>
<p>인터폰을 통해 호수의 호출을 눌러주세요.</p>
<p>예) 101호의 경우, "101" 입력 후 "호출" 클릭</p>
<p>호출 대기음이 들리면, 아래 버튼을 눌러주세요.</p>

<input id="pressing_button" style="height:90px;width:100%;font-size:40px;font-weight: bold;" type="submit" value="호출 수락하기" onclick="submit()">

<script type="text/javascript">
    function submit() {
        var xhr = new XMLHttpRequest();
        xhr.onreadystatechange = function () {
            if (xhr.readyState === 4) {
				document.getElementById("pressing_button").disabled = false;
				document.getElementById("pressing_button").value = "호출 수락하기";
                alert('완료되었습니다.');
            } else {
				document.getElementById("pressing_button").disabled = true;
				document.getElementById("pressing_button").value = "호출 수락 중...";
			}
        }
        xhr.open('GET', 'press', true);
        xhr.send();
    }

</script>
</body>
</html>
`)
)
