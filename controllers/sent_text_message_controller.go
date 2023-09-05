package controllers

import (
	"context"
	"zapmeow/models"
	"zapmeow/services"
	"zapmeow/utils"

	"github.com/gin-gonic/gin"
	waProto "go.mau.fi/whatsmeow/binary/proto"
)

type textMessageBody struct {
	Phone string
	Text  string
}

type sendTextMessageController struct {
	wppService     services.WppService
	messageService services.MessageService
}

func NewSendTextMessageController(
	wppService services.WppService,
	messageService services.MessageService,
) *sendTextMessageController {
	return &sendTextMessageController{
		wppService:     wppService,
		messageService: messageService,
	}
}

// Send Text Message on WhatsApp
// @Summary Send Text Message on WhatsApp
// @Description Sends a text message on WhatsApp using the specified instance.
// @Tags WhatsApp Chat
// @Param instanceId path string true "Instance ID"
// @Param data body textMessageBody true "Text message body"
// @Accept json
// @Produce json
// @Success 200 {object} string "Message Send Response"
// @Router /{instanceId}/chat/send/text [post]
func (t *sendTextMessageController) Handler(c *gin.Context) {
	var body textMessageBody
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.RespondBadRequest(c, "Body data is invalid")
		return
	}

	jid, ok := utils.MakeJID(body.Phone)
	if !ok {
		utils.RespondBadRequest(c, "Invalid phone")
		return
	}
	instanceId := c.Param("instanceId")

	instance, err := t.wppService.GetAuthenticatedInstance(instanceId)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: &body.Text,
		},
	}

	resp, err := instance.Client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	message := models.Message{
		ChatJID:    jid.User,
		SenderJID:  instance.Client.Store.ID.User,
		InstanceID: instance.ID,
		Body:       body.Text,
		Timestamp:  resp.Timestamp,
		FromMe:     true,
		MessageID:  resp.ID,
	}

	err = t.messageService.CreateMessage(&message)
	if err != nil {
		utils.RespondInternalServerError(c, err.Error())
		return
	}

	utils.RespondWithSuccess(c, gin.H{
		"Message": t.messageService.ToJSON(message),
	})
}
