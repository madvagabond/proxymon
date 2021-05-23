package speedtest
import (
	"testing"
	"fmt"
	"net/http"
)




func newtestClient() (Client, error) {
	return NewClient(&http.Client{})
}



func TestDownload(t *testing.T) {

	cli, e := newtestClient()

	if e != nil {
		t.Error("unable to initialize client")
	}

	speed, e := cli.DownloadTest()


	if e != nil {
		t.Error(e)
	}


	fmt.Println(speed)
}





func TestUpload(t *testing.T) {
	cli, e := newtestClient()

	if e != nil {
		t.Error("unable to initialize client")
	}

	speed, e := cli.UploadTest()
	

	if e != nil {
		t.Error(e)
	}


	fmt.Println(speed)
	
}



func TestPing(t *testing.T ) {
	cli, _ := newtestClient()
	latency, e := cli.PingTest()

	if e != nil {
		t.Error(e)
	}

	fmt.Println(latency)
}


func TestSpeedTest(t *testing.T) {
	cli, _ := newtestClient()
	result, e := cli.SpeedTest()

	if e != nil {
		t.Error(e)
	}

	fmt.Println(result)
	
}
