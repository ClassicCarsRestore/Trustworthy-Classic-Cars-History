export class User {
    constructor(
        public email: string,
        public role: string,
        public firstname: string,
        public surname: string,
        public country: string
    ) {}
}

export class UserLevel {
    constructor(
        public type: string,
    ) {}
}