import { Component, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import {User} from '../../user';
import { Router, ActivatedRoute} from '@angular/router';
import {Token} from '../../tokens';

@Component({
  selector: 'app-user-info',
  templateUrl: './user-info.component.html',
  styleUrls: ['./user-info.component.css']
})
export class UserInfoComponent {
  public user?: User;

  constructor(private http: HttpClient, private router: Router, @Inject('BASE_URL') private baseUrl: string, private activatedRoute: ActivatedRoute) {
    let email: string = "";

    // get id from url
    this.activatedRoute.paramMap.subscribe(params => {
      let tmpEmail: string | null = params.get('email');
      if (tmpEmail != null)
        email = tmpEmail;
    });
    
    http.get<User>(baseUrl + 'api/Users/Get/' + email, Token.getHeader()).subscribe(result => {
      this.user = result;
      console.log(this.user)
    }, error => {
      if (error && error.error && (error.error.message === "User with email "+email+" is not registered in the system.")) {
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
