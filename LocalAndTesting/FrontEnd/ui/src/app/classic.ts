export class Classic {
    constructor(
        public make: string,
        public model: string,
        public year: number,
        public licencePlate: string,
        public country: string,
        public chassisNo: string,
        public engineNo: string,
        public ownerEmail: string,
        public priorOwners: string[],
        public restorations: Restoration[],
        public certifications: string[],
        public documents: { [key: string]: string }
    ) {}
}

export class Restoration {
    constructor(
        public id: string,
        public title: string,
        public description: string,
        public photosIds: string[],
        public madeBy: string,
        public when: string,
    ) {}

}

export class ClassicForm {
    constructor(
        public make: string,
        public model: string,
        public year: number,
        public licencePlate: string,
        public country: string,
        public chassisNo: string,
        public engineNo: string,
        public ownerEmail: string,
    ) {}
}

export class ClassicEdit {
    constructor(
        public make: string,
        public model: string,
        public year: number,
        public licencePlate: string,
        public country: string,
        public engineNo: string,
    ) {}
}

export class ClassicAccess {
    constructor(
        public ownerEmail: string,
        public viewers: { [key: string]: string },
        public modifiers: { [key: string]: string },
        public certifiers: { [key: string]: string },
    ) {}
}

export class AccessBody {
    constructor(
        public email: string,
        public level: string,
    ) {}
}

export class ClassicHistory {
    constructor(
        public counter: number,
        public txns: ClassicTransaction[],
    ) {}
}

export class ClassicTransaction{
    constructor(
        public txn: string,
        public timestamp: string,
        public value: Classic,
    ) {}
}

export class AccessHistory {
    constructor(
        public counter: number,
        public txns: AccessTransaction[],
    ) {}
}

export class AccessTransaction{
    constructor(
        public txn: string,
        public timestamp: string,
        public value: ClassicAccess,
    ) {}
}

