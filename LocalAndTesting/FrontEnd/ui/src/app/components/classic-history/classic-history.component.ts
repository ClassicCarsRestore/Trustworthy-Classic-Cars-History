import { Component, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import {Token} from '../../tokens';
import { Router, ActivatedRoute } from '@angular/router';
import {ClassicHistory, Classic, ClassicTransaction} from '../../classic';
import { DatePipe } from '@angular/common';


@Component({
  selector: 'app-classic-history',
  templateUrl: './classic-history.component.html',
  styleUrls: ['./classic-history.component.css'],
})
export class ClassicHistoryComponent {
  public chassisNo?: string;
  public classic?: Classic;
  public classicHistory?: ClassicHistory;
  public txns?: ClassicTransaction[];
  public showPhotosMatrix: boolean[][] = [];
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
    http.get<ClassicHistory>(baseUrl+'api/Classics/History/' + this.chassisNo, Token.getHeader()).subscribe(result => {
      this.classicHistory = result;
      this.txns = result.txns;
      this.txns.forEach(() => {
        const row: boolean[] = new Array(500).fill(false);
        this.showPhotosMatrix.push(row);
      });
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

  showPhotos(i: number, j: number): void {
    this.showPhotosMatrix[i][j] = !this.showPhotosMatrix[i][j];
  }

  getDocuments( result: { [key: string]: string; }): [string, string][] {
    const documents = new Map<string, string>(Object.entries(result));
    const documentsArray = Array.from(documents.entries());
    return documentsArray;
  }

  getInfoByEmail(email: string) {
    this.router.navigate(['/users/'+email]);
  }
  
  handleImageError(event: Event) {
    const imgElement = event.target as HTMLImageElement;
    imgElement.src = 'assets/img/doc2.webp';
  }

}
