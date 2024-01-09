'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class MyWorkload extends WorkloadModuleBase {
    constructor() {
        super();
    }

    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);
    }

    async submitTransaction() {


        let requestSettings = {
            contractId: 'classiccars',
            contractFunction: 'GetAllClassics',
            invokerIdentity: 'org1-admin-default',
            timeout: 10,
            readOnly: true
        };
        
        await this.sutAdapter.sendRequests(requestSettings);
    }

    async cleanupWorkloadModule() {
    }
}

function createWorkloadModule() {
    return new MyWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;