import { Component, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import {Token} from '../../tokens';
import { Router, ActivatedRoute } from '@angular/router';
import {Classic} from '../../classic';
import { DatePipe } from '@angular/common';
import { Transaction } from 'src/app/transaction';

@Component({
  selector: 'app-transaction-info',
  templateUrl: './transaction-info.component.html',
  styleUrls: ['./transaction-info.component.css']
})
export class TransactionInfoComponent {
  public chassisNo?: string;
  public txId?: string;
  public classic?: Classic;
  public transaction?: Transaction;
  showDetails: boolean = false;
  showDetails1: boolean = false;

  constructor(private http: HttpClient, private router: Router, private activatedRoute: ActivatedRoute, @Inject('BASE_URL') private baseUrl: string,  private datePipe: DatePipe) {
    this.activatedRoute.paramMap.subscribe(params => {
      let tmpChassisNo: string | null = params.get('chassisNo');
      if (tmpChassisNo != null)
        this.chassisNo = tmpChassisNo;
      let tmpTxId: string | null = params.get('txId');
      if (tmpTxId != null)
          this.txId = tmpTxId;
    });
    http.get<Transaction>(baseUrl+'api/Transaction/' + this.chassisNo+'/'+this.txId, Token.getHeader()).subscribe(result => {
      this.transaction = result;
      console.log(this.transaction);
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

  getInfoByEmail(email: string) {
    this.router.navigate(['/users/'+email]);
  }

}
