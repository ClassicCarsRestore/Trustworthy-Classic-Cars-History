import { Component, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import {Token} from '../../tokens';
import {User, UserLevel} from '../../user';
import {ClassicAccess, AccessBody, Classic} from '../../classic';
import { Router, ActivatedRoute, NavigationExtras } from '@angular/router';
import { DatePipe } from '@angular/common';


@Component({
  selector: 'app-classic-update-email',
  templateUrl: './classic-update-email.component.html',
  styleUrls: ['./classic-update-email.component.css']
})
export class ClassicUpdateEmailComponent {
  public classic?: Classic;
  public isOwner: boolean = false;
  public users: User[];
  public chassisNo?: string;
  public showSpinner: boolean = false;
  public showSpinner2: boolean = false;
  public isConfirmed: boolean = false;
  public classicAccess?: ClassicAccess;
  public viewers: Map<string, string> = new Map<string, string>();
  public viewersArray: Array<[string, string]> = [];
  public modifiers: Map<string, string> = new Map<string, string>();
  public modifiersArray: Array<[string, string]> = [];
  public certifiers: Map<string, string> = new Map<string, string>();
  public certifiersArray: Array<[string, string]> = [];

  constructor(private http: HttpClient, private router: Router, private activatedRoute: ActivatedRoute, private datePipe: DatePipe, @Inject('BASE_URL') private baseUrl: string) {
    this.users = [];

    // get id from url
    this.activatedRoute.paramMap.subscribe(params => {
      let tmpChassisNo: string | null = params.get('chassisNo');
      if (tmpChassisNo != null)
        this.chassisNo = tmpChassisNo;
    });
    const navigation = this.router.getCurrentNavigation();
    if (navigation?.extras.state) {
      this.classic = navigation.extras.state['classic'];
    }
    else {
      http.get<Classic>(baseUrl+'api/Classics/Get/' + this.chassisNo, Token.getHeader()).subscribe(result => {
        this.classic = result;
      }, error => {
        if (error && error.error && (error.error.message === "404 - The classic "+this.chassisNo+" was Not Found" || error.error.message === "403 Forbidden - You are not authorized")) {
          alert(error.error.message);
          this.router.navigate(['/classics']);
        }
        console.log("An error occurred:", error);
      });
    }

    http.get<UserLevel>(baseUrl+'api/Access/CheckLevel/'+this.chassisNo, Token.getHeader()).subscribe(result => {
      if(result.type === "owner") {
        this.isOwner = true;
        http.get<ClassicAccess>(baseUrl+'api/Access/Get/'+this.chassisNo, Token.getHeader()).subscribe(result => {
          this.classicAccess = result;
          this.viewers = new Map<string, string>(Object.entries(result.viewers));
          this.viewersArray = Array.from(this.viewers.entries());
          this.modifiers = new Map<string, string>(Object.entries(result.modifiers));
          this.modifiersArray = Array.from(this.modifiers.entries());
          this.certifiers = new Map<string, string>(Object.entries(result.certifiers));
          this.certifiersArray = Array.from(this.certifiers.entries());
        }, error => {
          console.log("An error occurred:", error); 
        });
    
        http.get<User[]>(baseUrl+'api/Users/GetAll').subscribe(result => {
          this.users = result;
          console.log(this.users);
        }, error => {
          console.log("An error occurred:", error);
        });
      }
      else {
        alert("You do not have the necessary access level to access this page.")
        this.router.navigate(['/classics']);
      }
    }, error => {
      console.log("An error occurred:", error);
      this.router.navigate(['/classics']);
    });

  }

  onSubmitAccess() {
    let hasErrors: boolean = false;
    let grantOrRevoke = (<HTMLInputElement>document.getElementById("grantOrRevoke")).value;
    // disallow submitting without selecting an owner
    if (grantOrRevoke.trim() == "") {
      hasErrors = true;
      document.getElementById("grantOrRevokeEmpty")!.innerHTML = "You need to select Grant or Revoke";
    } else
      document.getElementById("grantOrRevokeEmpty")!.innerHTML = "";

    let typeOfAccess = (<HTMLInputElement>document.getElementById("typeOfAccess")).value;
    // disallow submitting without selecting an owner
    if (typeOfAccess.trim() == "") {
      hasErrors = true;
      document.getElementById("typeOfAccessEmpty")!.innerHTML = "You need to select one type of permission";
    } else
      document.getElementById("typeOfAccessEmpty")!.innerHTML = "";
    
    let userEmail = (<HTMLInputElement>document.getElementById("userEmail")).value;
    // disallow submitting without selecting an owner
    if (userEmail.trim() == "") {
      hasErrors = true;
      document.getElementById("userEmailEmpty")!.innerHTML = "An user's email needs to be selected";
    } else
      document.getElementById("userEmailEmpty")!.innerHTML = "";

    console.log(grantOrRevoke);
    console.log(typeOfAccess);
    console.log(userEmail);
    if (!hasErrors) {
      this.showSpinner2 = true;
      let body = new AccessBody(userEmail, typeOfAccess);
      if(grantOrRevoke === "grant") {
        this.http.put(this.baseUrl+'api/Access/Give/' + this.chassisNo, body, Token.getHeader()).subscribe(result => {
          this.navigateToSame(String(this.chassisNo));
        }, error => {
           if (error && error.error && (error.error.message === "404 - The classic "+this.chassisNo+" was Not Found" || error.error.message === "403 Forbidden - You are not authorized" ||error.error.message === "User with email "+userEmail+" is not registered in the system.")) {
             alert(error.error.message);
             this.router.navigate(['/classics']);
           }
          console.log("An error occurred:", error);
        });
      }
      else if(grantOrRevoke === "revoke") {
        this.http.put(this.baseUrl+'api/Access/Revoke/' + this.chassisNo, body, Token.getHeader()).subscribe(result => {
          this.navigateToSame(String(this.chassisNo));
        }, error => {
           if (error && error.error && (error.error.message === "404 - The classic "+this.chassisNo+" was Not Found" || error.error.message === "403 Forbidden - You are not authorized" ||error.error.message === "User with email "+userEmail+" is not registered in the system.")) {
             alert(error.error.message);
             this.router.navigate(['/classics']);
           }
          console.log("An error occurred:", error);
        });
      }
    }
  }

  navigateToSame(chassisNo: string) {
    const navigationExtras: NavigationExtras = {
      skipLocationChange: true // Navigate without changing the URL
    };
  
    this.router.navigate(['/classics/access/', chassisNo], navigationExtras).then(() => {
      window.location.reload();
    });
  }

  onSubmit() {
    let hasErrors: boolean = false;
    
    let newOwner = (<HTMLInputElement>document.getElementById("newOwner")).value;
    if (newOwner.trim() == "") {
      hasErrors = true;
      document.getElementById("newOwnerEmpty")!.innerHTML = "A new owner's email needs to be selected";
    } else
      document.getElementById("newOwnerEmpty")!.innerHTML = "";

    if (!hasErrors) {
      this.showSpinner = true;
      this.http.put(this.baseUrl+'api/Classics/UpdateEmail/' + this.chassisNo + "/" +newOwner, "", Token.getHeader()).subscribe(result => {
        this.router.navigate(['/classics']);
      }, error => {
         if (error && error.error && (error.error.message === "404 - The classic "+this.chassisNo+" was Not Found" || error.error.message === "403 Forbidden - You are not authorized" ||error.error.message === "User with email "+newOwner+" is not registered in the system.")) {
           alert(error.error.message);
           this.router.navigate(['/classics']);
         }
        console.log("An error occurred:", error);
      });
    }
  }

  formatDate(date: string): string {;
    return this.datePipe.transform(date, 'dd-MM-y, HH:mm') || '';
  }

  getInfoByEmail(email: string) {
    this.router.navigate(['/users/'+email]);
  }
}
