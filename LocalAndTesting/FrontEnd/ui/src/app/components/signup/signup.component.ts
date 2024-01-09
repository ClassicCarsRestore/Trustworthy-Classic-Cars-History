import { Component } from '@angular/core';
import { AuthService } from '../../services/auth.service';
import { CredentialsSignUp } from '../../credentials';
import { Router } from '@angular/router';
import { countries } from 'src/app/country-data-store';

@Component({
  selector: 'app-signup',
  templateUrl: './signup.component.html',
  styleUrls: ['./signup.component.css']
})
export class SignupComponent {
  public isConfirmed: boolean = false;
  public countries:any = countries;

  constructor(private router: Router, private authService: AuthService) {
    if(this.authService.isLoggedIn()) {
      this.router.navigate(['/classics/']);
    }
  }

  register(email: string, password: string, repeatPass: string, firstname: string, surname: string) {

    let hasErrors: boolean = false;
    if (firstname == '') {
      hasErrors = true;
      document.getElementById('firstnameEmpty')!.innerHTML = 'Invalid First Name!';
    } else {
      document.getElementById('firstnameEmpty')!.innerHTML = '';
    }
    if (surname == '') {
      hasErrors = true;
      document.getElementById('surnameEmpty')!.innerHTML = 'Invalid Surname!';
    } else {
      document.getElementById('surnameEmpty')!.innerHTML = '';
    }
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

    if (repeatPass == '') {
      hasErrors = true;
      document.getElementById('repeatPassEmpty')!.innerHTML = 'Invalid Password!';
    } else {
      document.getElementById('repeatPassEmpty')!.innerHTML = '';
    }
    let org = (<HTMLInputElement>document.getElementById("Organization")).value;
    let country = (<HTMLInputElement>document.getElementById("country")).value;
    if (country == '') {
      hasErrors = true;
      document.getElementById('countryEmpty')!.innerHTML = 'Invalid Country!';
    } else {
      document.getElementById('countryEmpty')!.innerHTML = '';
    }
    console.log(country);
    if (!hasErrors) {
      this.authService.register(new CredentialsSignUp(email, password, org, firstname, surname, country), repeatPass);
    }
  }

}
