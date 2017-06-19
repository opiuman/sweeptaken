// simple client which will push the messages to logrus
package kafka_logrus

/*
import (
	"bufio"
	"fmt"
	log "github.com/Sirupsen/logrus"
	kafkahook "github.com/goibibo/KafkaLogrus"
	"os"
)

func main() {
	addrs := []string{"127.0.0.1:9092"}
	hook, err := kafkahook.NewHook(addrs, "test_topic")
	if err == nil {
		log.AddHook(hook)
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter text: ")
	text, _ := reader.ReadString('\n')
	fmt.Println(text)
	log.Info(text)
}
*/
