import { Component, Inject } from '@angular/core';
import { Router } from '@angular/router';
import { HttpClient } from '@angular/common/http';
import {Token} from '../../tokens';
import {ClassicForm} from '../../classic';
import {User} from '../../user';
import { countries } from 'src/app/country-data-store';

@Component({
  selector: 'app-classic-new',
  templateUrl: './classic-new.component.html',
  styleUrls: ['./classic-new.component.css']
})
export class ClassicNewComponent {
  public users: User[];
  public showSpinner = false;
  public isConfirmed: boolean = false;
  public countries:any = countries;

  constructor(private http: HttpClient, private router: Router, @Inject('BASE_URL') private baseUrl: string) {
    this.users = [];
    http.get<User[]>(baseUrl+'api/Users/GetAll').subscribe(result => {
      this.users = result;
    }, error => {
      console.log("An error occurred:", error);
    });
  }

  onSubmit() {
    let hasErrors: boolean = false;

    let make = (<HTMLInputElement>document.getElementById("make")).value;
    if (make.trim() == "") {
      hasErrors = true;
      document.getElementById("makeEmpty")!.innerHTML = "Classic's make can't be empty";
    } else {
      document.getElementById("makeEmpty")!.innerHTML = "";
    }

    let model = (<HTMLInputElement>document.getElementById("model")).value;
    if (model.trim() == "") {
      hasErrors = true;
      document.getElementById("modelEmpty")!.innerHTML = "Classic's model can't be empty";
    } else {
      document.getElementById("modelEmpty")!.innerHTML = "";
    }

    let year = Number((<HTMLInputElement>document.getElementById("year")).value);
    if (isNaN(year) || year <= 0) {
      hasErrors = true;
      document.getElementById("yearEmpty")!.innerHTML = "Classic's year of manifacture must be a positive number";
    } else {
      document.getElementById("yearEmpty")!.innerHTML = "";
    }

    let licencePlate = (<HTMLInputElement>document.getElementById("licencePlate")).value;
    if (licencePlate.trim() == "") {
      hasErrors = true;
      document.getElementById("licencePlateEmpty")!.innerHTML = "Classic's license plate can't be empty";
    } else {
      document.getElementById("licencePlateEmpty")!.innerHTML = "";
    }

    let country = (<HTMLInputElement>document.getElementById("country")).value;
    if (country == '') {
      hasErrors = true;
      document.getElementById('countryEmpty')!.innerHTML = "Classic's country can't be empty";
    } else {
      document.getElementById('countryEmpty')!.innerHTML = '';
    }

    let chassisNo = (<HTMLInputElement>document.getElementById("chassisNo")).value;
    if (chassisNo.trim() == "") {
      hasErrors = true;
      document.getElementById("chassisNoEmpty")!.innerHTML = "Classic's chassis number/VIN can't be empty";
    } else {
      document.getElementById("chassisNoEmpty")!.innerHTML = "";
    }

    let engineNo = (<HTMLInputElement>document.getElementById("engineNo")).value;
    if (engineNo.trim() == "") {
      hasErrors = true;
      document.getElementById("engineNoEmpty")!.innerHTML = "Classic's engine details can't be empty";
    } else {
      document.getElementById("engineNoEmpty")!.innerHTML = "";
    }

    let ownerEmail = (<HTMLInputElement>document.getElementById("ownerEmail")).value;
    if (ownerEmail.trim() == "") {
      hasErrors = true;
      document.getElementById("ownerEmailEmpty")!.innerHTML = "A owner's email needs to be selected";
    } else {
      document.getElementById("ownerEmailEmpty")!.innerHTML = "";
    }

    if(!hasErrors) {
      this.showSpinner = true;
      let classic = new ClassicForm(make, model, year, licencePlate, country, chassisNo, engineNo, ownerEmail);
      this.http.post(this.baseUrl+'api/Classics/Create', classic, Token.getHeader()).subscribe(result => {
        alert("New Classic create successfully.")
        this.router.navigate(['/classics']);
      }, error => {
        if (error && error.error && (error.status === 409 || error.status === 400 || error.status === 403)) {
          alert(error.error.message);
          this.router.navigate(['/classics']);
        }
        console.log("Error status:", error.status); // Access error status
      });
    }
    
  }

}
