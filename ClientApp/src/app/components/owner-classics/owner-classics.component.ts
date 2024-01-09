import { Component, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import {Classic} from '../../classic';
import { Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import {Token} from '../../tokens';

@Component({
  selector: 'app-owner-classics',
  templateUrl: './owner-classics.component.html',
  styleUrls: ['./owner-classics.component.css']
})
export class OwnerClassicsComponent {
  public classics: Classic[];

  constructor(http: HttpClient, private router: Router, private authService: AuthService,  @Inject('BASE_URL') private baseUrl: string,) {
    this.classics = [];
    if(localStorage.getItem("loggedInOrg") === "Org1") {
      http.get<Classic[]>(baseUrl+'api/Classics/GetByOwner/', Token.getHeader()).subscribe(result => {
        this.classics = result;
      }, error => {
        console.log("An error occurred:", error);    
      });
    }
    else if(localStorage.getItem("loggedInOrg") === "Org2") {
      http.get<Classic[]>(baseUrl+'api/Classics/GetByModifier/', Token.getHeader()).subscribe(result => {
        this.classics = result;
      }, error => {
        console.log("An error occurred:", error);    
      });
    }
    else if(localStorage.getItem("loggedInOrg") === "Org3") {
      http.get<Classic[]>(baseUrl+'api/Classics/GetByCertifier/', Token.getHeader()).subscribe(result => {
        this.classics = result;
      }, error => {
        console.log("An error occurred:", error);  
      });
    }
  }

  getMapSize(map: any): number {
    return map ? Object.keys(map).length : 0;
  }

  goToClassicDetails(chassisNo: string) {
    this.router.navigate(['/classics/details/'+ chassisNo]);
  }
  
  isClient(): boolean {
    if(localStorage.getItem("loggedInOrg") === "Org1") {
      return true;
    }
    return false;
  }

  isWorkshop(): boolean {
    if(localStorage.getItem("loggedInOrg") === "Org2") {
      return true;
    }
    return false;
  }

  isCertifier(): boolean {
    if(localStorage.getItem("loggedInOrg") === "Org3") {
      return true;
    }
    return false;
  }
}
