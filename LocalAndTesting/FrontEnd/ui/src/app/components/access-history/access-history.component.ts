import { Component, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import {Token} from '../../tokens';
import { Router, ActivatedRoute } from '@angular/router';
import {AccessHistory, Classic, AccessTransaction} from '../../classic';
import { DatePipe } from '@angular/common';

@Component({
  selector: 'app-access-history',
  templateUrl: './access-history.component.html',
  styleUrls: ['./access-history.component.css']
})
export class AccessHistoryComponent {
  public chassisNo?: string;
  public classic?: Classic;
  public accessHistory?: AccessHistory;
  public txns?: AccessTransaction[];
  active = 1;

  constructor(private http: HttpClient, private router: Router, private activatedRoute: ActivatedRoute, @Inject('BASE_URL') private baseUrl: string,  private datePipe: DatePipe) {
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
    http.get<AccessHistory>(baseUrl+'api/Acess/History/' + this.chassisNo, Token.getHeader()).subscribe(result => {
      this.accessHistory = result;
      this.txns = result.txns;
      console.log(this.accessHistory);
    }, error => {
      if (error && error.error && (error.error.message === "404 - The classic "+this.chassisNo+" was Not Found" || error.error.message === "403 Forbidden - You are not authorized")) {
        alert(error.error.message);
        this.router.navigate(['/classics']);
      }
      console.log("An error occurred:", error);
    });
  }

  formatDate(date: string): string {
    return this.datePipe.transform(date, 'dd-MM-y, HH:mm') || '';
  }

  getResults( result: { [key: string]: string; }): [string, string][] {
    const results = new Map<string, string>(Object.entries(result));
    const resultsArray = Array.from(results.entries());
    return resultsArray;
  }

  getInfoByEmail(email: string) {
    this.router.navigate(['/users/'+email]);
  }

}


