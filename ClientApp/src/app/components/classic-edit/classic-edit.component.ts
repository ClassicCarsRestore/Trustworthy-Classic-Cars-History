import { Component, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import {Classic, ClassicEdit} from '../../classic';
import { Router, ActivatedRoute } from '@angular/router';
import {Token} from '../../tokens';
import { countries } from 'src/app/country-data-store';

@Component({
  selector: 'app-classic-edit',
  templateUrl: './classic-edit.component.html',
  styleUrls: ['./classic-edit.component.css']
})
export class ClassicEditComponent{ 
  public classic?: Classic;
  public showSpinner = false;
  public countries:any = countries;

  constructor(private http: HttpClient, private router: Router, private activatedRoute: ActivatedRoute, @Inject('BASE_URL') private baseUrl: string) {
    const navigation = this.router.getCurrentNavigation();
    if (navigation?.extras.state) {
      this.classic = navigation.extras.state['classic'];
    }
    else {
      let chassisNo: string = "";
  
      // get id from url
      this.activatedRoute.paramMap.subscribe(params => {
        let tmpChassisNo: string | null = params.get('chassisNo');
        if (tmpChassisNo != null)
          chassisNo = tmpChassisNo;
      });
  
      http.get<Classic>(baseUrl+'api/Classics/Get/' + chassisNo, Token.getHeader()).subscribe(result => {
        this.classic = result;
        console.log(this.classic)
      }, error => {
        if (error && error.error && (error.error.message === "404 - The classic "+chassisNo+" was Not Found" || error.error.message === "403 Forbidden - You are not authorized")) {
          alert(error.error.message);
          this.router.navigate(['/classics']);
        }
        console.log("An error occurred:", error);  
      });
    }
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

    let engineNo = (<HTMLInputElement>document.getElementById("engineNo")).value;
    if (engineNo.trim() == "") {
      hasErrors = true;
      document.getElementById("engineNoEmpty")!.innerHTML = "Classic's engine details can't be empty";
    } else {
      document.getElementById("engineNoEmpty")!.innerHTML = "";
    }

    if (!hasErrors && this.classic) {
      this.showSpinner = true;
      let body = new ClassicEdit(make, model, year, licencePlate, country, engineNo);
      this.http.put(this.baseUrl+'api/Classics/Update/' + this.classic.chassisNo, body, Token.getHeader()).subscribe(result => {
        console.log(result);
        this.router.navigate(['/classics/details/'+ this.classic?.chassisNo]);
      }, error => {
        if (error && error.error && (error.error.message === "404 - The classic "+this.classic?.chassisNo+" was Not Found" || error.error.message === "403 Forbidden - You are not authorized")) {
          alert(error.error.message);
          this.router.navigate(['/classics']);
        }
        console.log("An error occurred:", error);
        console.log("Error status:", error.status);
        console.log("Error message:", error.error.message);
        console.log("Error success:", error.error.success); 
      });
    }
  }
}
