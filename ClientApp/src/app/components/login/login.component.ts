import { Component } from '@angular/core';
import { AuthService } from '../../services/auth.service';
import { Credentials } from '../../credentials';
import { Router } from '@angular/router';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css'],
})
export class LoginComponent {

  constructor(private router: Router, private authService: AuthService) {
    if(this.authService.isLoggedIn()) {
      this.router.navigate(['/classics/']);
    }
  }

  login(email: string, password: string) {

    let hasErrors: boolean = false;

    if (email == '') {
      hasErrors = true;
      document.getElementById('emailEmpty')!.innerHTML = 'Invalid Email!';
    } else {
      document.getElementById('emailEmpty')!.innerHTML = '';
    }
    if (password == '') {
      hasErrors = true;
      document.getElementById('passwordEmpty')!.innerHTML = 'Invalid Password!';
    } else {
      document.getElementById('passwordEmpty')!.innerHTML = '';
    }
    let org = (<HTMLInputElement>document.getElementById("Organization")).value;
    
    if (!hasErrors) {
      this.authService.login(new Credentials(email, password, org));
    }
  }
}
