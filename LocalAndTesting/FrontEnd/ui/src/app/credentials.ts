export class Credentials {
  constructor(
    public email: string,
    public password: string,
    public orgname: string
  ) {}
}

export class CredentialsSignUp {
  constructor(
    public email: string,
    public password: string,
    public orgname: string,
    public firstname: string,
    public surname: string,
    public country: string
  ) {}
}

export class CredentialToken {
  constructor(public success: string, public message: Message) {}
}

export class Message {
  constructor(public email: string, public token: string, public org: string) {}
}
