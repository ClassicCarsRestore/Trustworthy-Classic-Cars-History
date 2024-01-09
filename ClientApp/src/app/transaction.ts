export class Transaction {
    constructor(
        public txId: string,
        public timestamp: string,
        public creator: Creator,
        public fx: Function,
        public response: Response,
    ) {}
}

export class Creator {
    constructor(
        public email: string,
        public orgMsp: string
    ) {}
}

export class Function {
    constructor(
        public name: string,
        public args: string[]
    ) {}
}

export class Response {
    constructor(
        public response: number,
        public result: string
    ) {}
}