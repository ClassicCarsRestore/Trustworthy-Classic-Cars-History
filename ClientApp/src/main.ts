/// <reference types="@angular/localize" />

import { platformBrowserDynamic } from '@angular/platform-browser-dynamic';

import { AppModule } from './app/app.module';

export function getBaseUrl() {
  return "https://gui.classicschain.com:8393/";
}

const providers = [
  { provide: 'BASE_URL', useFactory: getBaseUrl, deps: [] }
];


platformBrowserDynamic(providers).bootstrapModule(AppModule)
  .catch(err => console.error(err));
