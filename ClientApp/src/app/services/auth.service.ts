import { Injectable, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject } from 'rxjs';
import { Credentials, CredentialToken, CredentialsSignUp } from '../credentials';
import { Router, ActivatedRoute } from '@angular/router';

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private email: BehaviorSubject<string> = new BehaviorSubject<string>("");
  private org: BehaviorSubject<string> = new BehaviorSubject<string>("");

  get loggedInEmailObservable() {
    return this.email.asObservable();
  }

  get loggedInEmail() {
    return this.email.value;
  }

  get loggedInOrgObservable() {
    return this.org.asObservable();
  }

  get loggedInOrg() {
    return this.org.value;
  }

  isLoggedIn(): boolean {
    const token = localStorage.getItem('ChainToken');

    if (token) {
        return true; // The user is considered logged in.
    } else {
        return false; // No token found, user is not logged in.
    }
}

  constructor(public client: HttpClient, private router: Router, @Inject('BASE_URL') private baseUrl: string, private route: ActivatedRoute) {}

  login(cred: Credentials) {
    if (cred.email !== '' && cred.password !== '' && cred.orgname !== '') {
      this.client
        .post<CredentialToken>(this.baseUrl+'api/Users/Login', cred)
        .subscribe(
          (result) => {
            console.log(result);
            console.log(result.message.email)
            console.log(result.message.token)
            localStorage.setItem('ChainToken', result.message.token);
            localStorage.setItem('loggedInEmail', result.message.email);
            localStorage.setItem('loggedInOrg', result.message.org);
            this.email.next(result.message.email);
            this.org.next(cred.orgname);
            const returnUrl = this.route.snapshot.queryParams['returnUrl'] || '/classics';
            console.log(returnUrl);
            this.router.navigateByUrl(returnUrl);
          },
          (error) => {
            console.log("An error occurred:", error);
            console.log("Error status:", error.status); 
            console.log("Error message:", error.error.message);
            console.log("Error success:", error.error.success);
            document.getElementById("passwordEmpty")!.innerHTML = error.error.message;
          }
        );
    }
  }

  logout() {
    localStorage.removeItem('ChainToken');
    localStorage.removeItem('loggedInEmail');
    localStorage.removeItem('loggedInOrg');
    this.email.next("");
    this.org.next("");
    this.router.navigate(['/login']);
  }

  register(cred: CredentialsSignUp, repeatPass: string) {
    if (cred.email !== '' && cred.password !== '' && repeatPass !== '' && cred.orgname !== '' && cred.firstname !== '' && cred.surname !== '' && cred.country !== '') {
      if(cred.password.trim() !== repeatPass.trim()) {
        document.getElementById("repeatPassEmpty")!.innerHTML = "The passwords do not match.";
      }
      else {
        this.client
          .post(this.baseUrl+'api/Users', cred)
          .subscribe(
            (result) => {
              alert("User registered succesfully. Welcome to our platform.")
              this.router.navigate(['/login']);
            },
            (error) => {
              console.log("An error occurred:", error);
              document.getElementById("repeatPassEmpty")!.innerHTML = error.error.message;
            }
          );
      }
      }

  }
}
