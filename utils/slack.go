package slack

import (
	"encoding/json"
	"fmt"
	"github.com/slack-go/slack"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type slackMessage struct {
	Name      string       `json:"name"`
	Namespace string       `json:"namespace"`
	Phase     string       `json:"phase"`
	StartTime *metav1.Time `json:"startTime"`
}

func BuildSlackAttachment(eventType string, pod *corev1.Pod, restartCount int) slack.Attachment {

	var msg string
	color := "#36a64f" // green by default

	switch eventType {
	case "CrashLoopBackOff":
		msg = fmt.Sprintf("🔥 CrashLoopBackOff pod detected: %s/%s (restarts: %d)", pod.Namespace, pod.Name, restartCount)
		color = "#ffae42" // orange

	case "FailedOrEvicted":
		msg = fmt.Sprintf("🔥 Failed/Evicted pod detected: %s/%s", pod.Namespace, pod.Name)
		color = "#d72d2d" // red
	case "FailedToDelete":
		msg = fmt.Sprintf("🔥 Failed to delete pod detected: %s/%s", pod.Namespace, pod.Name)
		color = "#ffff00" // red

	case "Deleted":
		obj := slackMessage{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Phase:     string(pod.Status.Phase),
			StartTime: pod.Status.StartTime,
		}
		b, _ := json.MarshalIndent(obj, "", "  ")
		msg = fmt.Sprintf("✅ Deleted pod:\n```%s```", string(b))
		color = "#2eb886" // green
	}

	return slack.Attachment{
		Pretext: "Pod Cleanup Controller Notification",
		Text:    msg,
		Color:   color,
		Fields: []slack.AttachmentField{
			{
				Title: "Timestamp",
				Value: time.Now().Format(time.RFC3339),
				Short: false,
			},
		},
	}
}

/*
attachment := buildSlackAttachment("CrashLoopBackOff", pod, 21)
slackClient.PostMessage(channelID, slack.MsgOptionAttachments(attachment))


attachment := buildSlackAttachment("Deleted", pod, 0)
slackClient.PostMessage(channelID, slack.MsgOptionAttachments(attachment))


**/
