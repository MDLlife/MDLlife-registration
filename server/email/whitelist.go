package email

import (
	"../ses"
	"../config"
)

func ConfirmEmail(to string, token string) {
	emailData := ses.Email{
	To:   to,
	From: config.Config.NoReplyEmail,
	Text: "Your whitelist submission is well received.\n\n" +
	"To finish the whitelist application process please confirm your email by following the link/n" +
	"https://mdl.life/whitelist/confirm_email?token=" + token + "\n\n" +
	"The instructions of how to purchase the MDL Tokens to be send soon is confirmation that you have passed the whitelist.\n\n" +
	"For inquiries and support please contact support@mdl.life",
	HTML: "<h3 style=\"color:purple;\">Your whitelist submission is well received.</h3><br>" +
	"To finish the whitelist application process please confirm your email by clicking the link<br>" +
	"<a href=\"https://mdl.life/whitelist/confirm_email?token=" + token + "\">" + "https://mdl.life/whitelist/confirm_email?token=" + token + "</a><br><br>" +
	"The instructions of how to purchase the MDL Tokens to be send soon is confirmation that you have passed the whitelist.<br><br>" +
	"For inquiries and support please contact <a href=\"mailto:support@mdl.life\">support@mdl.life</a>",
	Subject: "MDL Talent Hub: Whitelist application received",
	ReplyTo: config.Config.ReplyEmail,
	}

	ses.SendEmail(emailData)
}
