# SMS Phone Verification in Go

This is a small Go app that shows how to perform SMS phone verification.

Find out more on [Twilio Code Exchange][code-exchange-url].

## Application Overview

It uses [Twilio Verify][twilio-verify-url] to verify phone numbers and add an additional layer of security, helping prevent fraudulent users from registering with your business.

It steps users through the process of creating an account on a (fictional) website; specifically:

1. The user needs to enter a username, password, and phone number in the initial step and submit the form.
1. Then, the customer is sent an SMS with a verification code for 2-Factor Authentication (2FA), and redirected to the second step in the process where they will be asked to enter a verification code into the webpage to verify their account.
1. The verification code is then validated against the Twilio Verify service.
  If the code is valid, the user is then redirected to the logged in page, indicating that they have successfully registered with the application.
  If the code is not valid, the user is redirected back to the verify step, where an error is displayed, showing that verification failed.

## Requirements

To use the application, you'll need the following:

- [Go][go-download-url] 1.22 or above
- A Twilio account (free or paid) with a phone number. [Click here to create one][twilio-referral-url], if you don't have one.
- A phone number that can received SMS
- Your web browser of choice

## Getting Started

After cloning the code to where you store your Go projects, and change into the project directory.
Then, copy _.env.example_ as _.env_, by running the following command:

```bash
cp -v .env.example .env
```

After that, set values for `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, `TWILIO_PHONE_NUMBER`.
You can retrieve these details from the **Account Info** panel of your [Twilio Console][twilio-console-url] dashboard.

![A screenshot of the Account Info panel in the Twilio Console dashboard. It shows three fields: Account SID, Auth Token, and "My Twilio phone number", where Account SID and "My Twilio phone number" are redacted.](docs/images/twilio-console-account-info-panel.png)

### Create a Verify Service

Following that, create a Verify V2 service.

![The Verify V2 services page in the Twilio Console](./docs/images/twilio-verify-services.png)

First, open [the Twilio Console][twilio-console-url] in your browser of choice and navigate to **Explore products > Verify >** [Services][twilio-console-verify-services-url].
There, click **Create new**.

![The initial form for creating a Verify V2 service in the Twilio Console](./docs/images/create-twilio-verify-service-step-one.png)

In the **_Create new_** (Verify Service) form that appears, provide a **Friendly name**, enable the **SMS** verification channel, and click **Continue**.

![The Enable Fraud Guard stage of creating a new Verify V2 service in the Twilio Console](./docs/images/create-twilio-verify-service-step-two.png)

Following that, click **Continue** in the **Enable Fraud Guard** stage.

![The settings page of a Verify V2 service in the Twilio Console](./docs/images/twilio-verify-service-settings.png)

Now, you'll be on the **Service** settings page for your new Verify Service.
Copy the **Service SID** and set it as the value of `TWILIO_VERIFICATION_SID` in _.env_.

### Launch the Application

When that's done, run the following command to start the application:

```php
go run main.go
```

With the application ready to go, open <http://localhost:8000> in your browser of choice.
There, you'll see the form where you can request a verification code. 

![A web-based form to request a verification code. The form has three fields sorted vertically, and a button to submit the form, labelled "Request Verification Code". The first field is for the user's username, which can be between five and 255 chars. The second is for the user's password, which must be at least 10 characters long. The third is for the user's phone number, which must be in E.164 format.](./docs/images/request-verification-code.png)

Enter your username, password, and phone number and submit the form.
Then, you'll see a form which you can validate the verification code.

![A web-based form to validate a verification code. The form has one field, which is to take the user's verification code, and a button to submit the form, labelled "Validate Verification Code".](./docs/images/validate-verification-code.png)

Enter the verification code that you should have received via SMS and submit the form.
You will now be "authenticated" with the application, and on the logged in page.

![A page showing that a user is logged in with the message "You are now logged in." in the center of the page](./docs/images/logged-in-page.png)

There is no explicit log out functionality. 
To do that, you'll have to clear the cookies for localhost using your browser's settings tooling.

## Contributing

If you want to contribute to the project, whether you have found issues with it or just want to improve it, here's how:

- [Issues][github-issues-url]: ask questions and submit your feature requests, bug reports, etc
- [Pull requests][github-pr-url]: send your improvements

## Resources

Find out more about the project on [CodeExchange][code-exchange-url].

## Did You Find The Project Useful?

If the project was useful and you want to say thank you and/or support its active development, here's how:

- Add a GitHub Star to the project
- Write an interesting article about the project wherever you blog

[code-exchange-url]: https://www.twilio.com/code-exchange/sms-phone-verification
[go-download-url]: https://go.dev/doc/install
[twilio-console-url]: https://console.twilio.com/
[twilio-referral-url]: http://www.twilio.com/referral/QlBtVJ
[twilio-verify-url]: https://www.twilio.com/docs/verify
[github-issues-url]: https://github.com/settermjd/sms-phone-verification-go/issues
[github-pr-url]: https://github.com/settermjd/sms-phone-verification-go/pulls
[twilio-console-verify-services-url]: https://console.twilio.com/us1/develop/verify/services
