package email

import "fmt"

// generateVerificationEmailHTML generates HTML email template for email verification
func generateVerificationEmailHTML(otpCode, userEmail string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Verifikasi Email - Bantuaku</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #0a0a0a; color: #e2e8f0;">
	<div style="max-width: 600px; margin: 0 auto; padding: 40px 20px; background-color: #0a0a0a;">
		<div style="text-align: center; margin-bottom: 40px;">
			<h1 style="color: #10b981; font-size: 32px; margin: 0; font-weight: 700; text-shadow: 0 0 20px rgba(16, 185, 129, 0.3);">
				Bantuaku
			</h1>
		</div>
		
		<div style="background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 12px; padding: 40px; backdrop-filter: blur(10px);">
			<h2 style="color: #e2e8f0; font-size: 24px; margin: 0 0 20px 0; font-weight: 600;">
				Verifikasi Email Anda
			</h2>
			
			<p style="color: #94a3b8; font-size: 16px; line-height: 1.6; margin: 0 0 30px 0;">
				Terima kasih telah mendaftar di Bantuaku! Untuk menyelesaikan pendaftaran, silakan verifikasi alamat email Anda dengan kode berikut:
			</p>
			
			<div style="text-align: center; margin: 40px 0;">
				<div style="display: inline-block; background: linear-gradient(135deg, #10b981 0%%, #059669 100%%); padding: 20px 40px; border-radius: 12px; box-shadow: 0 0 30px rgba(16, 185, 129, 0.4);">
					<div style="font-size: 48px; font-weight: 700; color: #ffffff; letter-spacing: 8px; font-family: 'Courier New', monospace;">
						%s
					</div>
				</div>
			</div>
			
			<p style="color: #94a3b8; font-size: 14px; line-height: 1.6; margin: 30px 0 0 0; text-align: center;">
				Kode ini akan kedaluwarsa dalam <strong style="color: #10b981;">1 jam</strong>.
			</p>
			
			<p style="color: #64748b; font-size: 12px; line-height: 1.6; margin: 40px 0 0 0; text-align: center; border-top: 1px solid rgba(255, 255, 255, 0.1); padding-top: 20px;">
				Jika Anda tidak meminta kode verifikasi ini, Anda dapat mengabaikan email ini.
			</p>
		</div>
		
		<div style="text-align: center; margin-top: 40px; padding-top: 20px; border-top: 1px solid rgba(255, 255, 255, 0.1);">
			<p style="color: #64748b; font-size: 12px; margin: 0;">
				&copy; 2025 Bantuaku. All rights reserved.
			</p>
		</div>
	</div>
</body>
</html>`, otpCode)
}

// generateVerificationEmailText generates plain text email for email verification
func generateVerificationEmailText(otpCode, userEmail string) string {
	return fmt.Sprintf(`Bantuaku - Verifikasi Email

Terima kasih telah mendaftar di Bantuaku!

Untuk menyelesaikan pendaftaran, silakan verifikasi alamat email Anda dengan kode berikut:

%s

Kode ini akan kedaluwarsa dalam 1 jam.

Jika Anda tidak meminta kode verifikasi ini, Anda dapat mengabaikan email ini.

---
© 2025 Bantuaku. All rights reserved.`, otpCode)
}

// generatePasswordResetEmailHTML generates HTML email template for password reset
func generatePasswordResetEmailHTML(resetLink, userEmail string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Reset Password - Bantuaku</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #0a0a0a; color: #e2e8f0;">
	<div style="max-width: 600px; margin: 0 auto; padding: 40px 20px; background-color: #0a0a0a;">
		<div style="text-align: center; margin-bottom: 40px;">
			<h1 style="color: #10b981; font-size: 32px; margin: 0; font-weight: 700; text-shadow: 0 0 20px rgba(16, 185, 129, 0.3);">
				Bantuaku
			</h1>
		</div>
		
		<div style="background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 12px; padding: 40px; backdrop-filter: blur(10px);">
			<h2 style="color: #e2e8f0; font-size: 24px; margin: 0 0 20px 0; font-weight: 600;">
				Reset Password
			</h2>
			
			<p style="color: #94a3b8; font-size: 16px; line-height: 1.6; margin: 0 0 30px 0;">
				Kami menerima permintaan untuk mereset password akun Bantuaku Anda. Klik tombol di bawah untuk membuat password baru:
			</p>
			
			<div style="text-align: center; margin: 40px 0;">
				<a href="%s" style="display: inline-block; background: linear-gradient(135deg, #10b981 0%%, #059669 100%%); color: #ffffff; text-decoration: none; padding: 16px 32px; border-radius: 8px; font-weight: 600; font-size: 16px; box-shadow: 0 0 20px rgba(16, 185, 129, 0.3); transition: all 0.3s;">
					Reset Password
				</a>
			</div>
			
			<p style="color: #94a3b8; font-size: 14px; line-height: 1.6; margin: 30px 0 0 0;">
				Atau salin dan tempel link berikut ke browser Anda:
			</p>
			<p style="color: #10b981; font-size: 12px; line-height: 1.6; margin: 10px 0 0 0; word-break: break-all; font-family: 'Courier New', monospace;">
				%s
			</p>
			
			<p style="color: #94a3b8; font-size: 14px; line-height: 1.6; margin: 30px 0 0 0; text-align: center;">
				Link ini akan kedaluwarsa dalam <strong style="color: #10b981;">1 jam</strong>.
			</p>
			
			<p style="color: #64748b; font-size: 12px; line-height: 1.6; margin: 40px 0 0 0; text-align: center; border-top: 1px solid rgba(255, 255, 255, 0.1); padding-top: 20px;">
				Jika Anda tidak meminta reset password, Anda dapat mengabaikan email ini. Password Anda tidak akan berubah.
			</p>
		</div>
		
		<div style="text-align: center; margin-top: 40px; padding-top: 20px; border-top: 1px solid rgba(255, 255, 255, 0.1);">
			<p style="color: #64748b; font-size: 12px; margin: 0;">
				&copy; 2025 Bantuaku. All rights reserved.
			</p>
		</div>
	</div>
</body>
</html>`, resetLink, resetLink)
}

// generatePasswordResetEmailText generates plain text email for password reset
func generatePasswordResetEmailText(resetLink, userEmail string) string {
	return fmt.Sprintf(`Bantuaku - Reset Password

Kami menerima permintaan untuk mereset password akun Bantuaku Anda.

Klik link berikut untuk membuat password baru:

%s

Link ini akan kedaluwarsa dalam 1 jam.

Jika Anda tidak meminta reset password, Anda dapat mengabaikan email ini. Password Anda tidak akan berubah.

---
© 2025 Bantuaku. All rights reserved.`, resetLink)
}

