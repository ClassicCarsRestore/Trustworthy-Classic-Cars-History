import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import {NgbModule, NgbNavModule } from '@ng-bootstrap/ng-bootstrap';

import { AppComponent } from './app.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { NavMenuComponent } from './components/nav-menu/nav-menu.component';
import { HttpClientModule, HTTP_INTERCEPTORS } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { RouterModule, Routes } from '@angular/router';
import { AboutComponent } from './components/about/about.component';
import { DatePipe } from '@angular/common';
import { MaterialModule } from '../material.module';
import { LoginComponent } from './components/login/login.component';
import { OwnerClassicsComponent } from './components/owner-classics/owner-classics.component';
import { HomeComponent } from './components/home/home.component';

import { TokenExpirationInterceptor } from './services/token-expiration-interceptor.service';
import { ClassicDetailsComponent } from './components/classic-details/classic-details.component';
import { ClassicEditComponent } from './components/classic-edit/classic-edit.component';
import { ClassicUpdateEmailComponent } from './components/classic-update-email/classic-update-email.component';
import { RestorationDetailsComponent } from './components/restoration-details/restoration-details.component';
import { RestorationCreateComponent } from './components/restoration-create/restoration-create.component';
import { DocumentFormComponent } from './components/document-form/document-form.component';
import { ClassicHistoryComponent } from './components/classic-history/classic-history.component';
import { AccessHistoryComponent } from './components/access-history/access-history.component';
import { ClassicNewComponent } from './components/classic-new/classic-new.component';
import { SignupComponent } from './components/signup/signup.component';
import { UserInfoComponent } from './components/user-info/user-info.component';


const appRoutes: Routes = [
  { path: '', component: HomeComponent },
  { path: 'about', component: AboutComponent },
  { path: 'login', component: LoginComponent },
  { path: 'signup', component: SignupComponent },
  { path: 'classics', component: OwnerClassicsComponent },
  { path: 'classics/details/:chassisNo', component: ClassicDetailsComponent },
  { path: 'classics/details/:chassisNo/edit', component: ClassicEditComponent },
  { path: 'classics/access/:chassisNo', component: ClassicUpdateEmailComponent },
  { path: 'restoration/details/:chassisNo/:stepId', component: RestorationDetailsComponent },
  { path: 'restorations/create/:chassisNo', component: RestorationCreateComponent },
  { path: 'classics/document/:chassisNo', component: DocumentFormComponent },
  { path: 'classics/history/:chassisNo', component: ClassicHistoryComponent },
  { path: 'classics/access/history/:chassisNo', component: AccessHistoryComponent },
  { path: 'classics/new', component: ClassicNewComponent },
  { path: 'users/:email', component: UserInfoComponent },
];

@NgModule({
  declarations: [
    AppComponent,
    NavMenuComponent,
    AboutComponent,
    LoginComponent,
    OwnerClassicsComponent,
    HomeComponent,
    ClassicDetailsComponent,
    ClassicEditComponent,
    ClassicUpdateEmailComponent,
    RestorationDetailsComponent,
    RestorationCreateComponent,
    DocumentFormComponent,
    ClassicHistoryComponent,
    AccessHistoryComponent,
    ClassicNewComponent,
    SignupComponent,
    UserInfoComponent,
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    HttpClientModule,
    FormsModule,
    // MatToolbarModule,
    // MatButtonModule,
    MaterialModule,
    NgbModule,
    NgbNavModule,
    RouterModule.forRoot(appRoutes, { enableTracing: true }),
  ],
  providers: [
    {
      provide: HTTP_INTERCEPTORS,
      useClass: TokenExpirationInterceptor,
      multi: true,
    },
    DatePipe
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
