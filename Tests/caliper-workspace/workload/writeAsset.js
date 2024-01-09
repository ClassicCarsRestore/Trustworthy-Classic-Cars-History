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

        let emptyPhotos = [];

        let requestSettings = {
            contractId: 'classiccars',
            contractFunction: 'CreateRestorationStep',
            invokerIdentity: 'org1-admin-default',
            args: ['99990', 'Step title', 'Step description', emptyPhotos, "testTime"]
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