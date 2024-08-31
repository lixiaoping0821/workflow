package node

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

// smtp服务器配置
type SmtpConfig struct {
	Host   string `json:"host"` //smtp.qq.com
	Port   int    `json:"port"` //465
	User   string `json:"user"`
	Passwd string `json:"passwd"` //fykestarxulabcjf
}

/* {"smtpConfig":{"host":"smtp.163.com","port":"25","user":"","passwd":"4444",
	},
	"from":"123@qq.com","to":"456@qq.com",
	"cc":"123@qq.com,123@qq.com","subject":"主题",
	"content":"邮件内容","attachments":["123","345"]
} */
// smtp发送邮件配置
type SmtpNode struct {
	SmtpConfig  SmtpConfig             `form:"smtpConfig" json:"smtpConfig"`
	From        string                 `form:"from" json:"from"`
	To          []string               `form:"to" json:"to"`
	Cc          []string               `form:"cc" json:"cc"`
	Subject     string                 `form:"subject" json:"subject"`
	Content     string                 `form:"content" json:"content"`
	Attachments []multipart.FileHeader `form:"attachments"`
}

var Smtp *SmtpNode

func init() {
	// 初始化smtp服务器的配置，可以从配置文件获取
	smtpConfig := SmtpConfig{Host: "smtp.qq.com", Port: 465, User: "995903799@qq.com", Passwd: "ldrsrttyoxttbbga"}
	Smtp = NewSmtpNode(smtpConfig)
	fmt.Printf("初始化了smtp配置%#v", Smtp)
}

func NewSmtpNode(smtpConfig SmtpConfig) *SmtpNode {
	return &SmtpNode{
		SmtpConfig: smtpConfig,
	}
}

func (smtp *SmtpNode) GetKind() string {
	return "smtp"
}

// 发送带附件的邮件
func (smtp *SmtpNode) SendEmailAttach(c *gin.Context) {
	if err := c.ShouldBind(&smtp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("smtp的配置信息和发送信息 %#v", smtp.To)
	// return
	smtpConfig := Smtp.SmtpConfig
	smtpHost := smtpConfig.Host
	smtpPort := smtpConfig.Port
	smtpUser := smtpConfig.User
	smtpPass := smtpConfig.Passwd

	from := smtp.From
	to := smtp.To
	if from == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "40001", "msg": "from:发件人不能为空"})
		return
	}
	if to == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "40002", "msg": "to[]:收件人不能为空"})
		return
	}
	subject := smtp.Subject
	if subject == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "40003", "msg": "subject:主题不能为空"})
		return
	}
	content := smtp.Content
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "40004", "msg": "content:邮件正文不能为空"})
		return
	}

	// 创建邮件消息
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeaders(map[string][]string{
		"To": smtp.To,
		"Cc": smtp.Cc,
	})
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", content)
	// 处理发送邮件的附件部分
	// 获取当前时间格式化字符串,微秒级，防止时间冲突
	timestamp := time.Now().Format("20060102150405.000000")
	// 创建文件夹
	dir := fmt.Sprintf("./smtpfile/%s", timestamp)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 添加附件
	for _, fileHeader := range smtp.Attachments {
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer file.Close()

		// 读取文件内容
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 文件路径
		filename := fileHeader.Filename
		filePath := fmt.Sprintf("%s/%s", dir, filename)

		// 写入文件
		if err := os.WriteFile(filePath, fileBytes, 0666); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 添加附件,发送完成之后记得删除附件
		msg.Attach(filePath)
		defer func() {
			// 直接删除目录，不管下面是否有文件
			if err := os.RemoveAll(dir); err != nil {
				log.Printf("Failed to remove file %s: %v", dir, err)
			}
		}()
	}
	// msg.Attach("./file/php.jpeg")
	// msg.Attach("./file/php.xlsx")

	// 设置SMTP服务器并发送邮件
	var d = gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)

	if err := d.DialAndSend(msg); err != nil {
		// 邮件发送成功后删除本地附件
		for _, fileHeader := range smtp.Attachments {
			filePath := fmt.Sprintf("%s/%s", dir, fileHeader.Filename)
			if err := os.Remove(filePath); err != nil {
				log.Printf("Failed to remove file %s: %v", filePath, err)
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": "40005", "发送邮件错误error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"code": "0", "message": "Email sent successfully"})
	}
}
