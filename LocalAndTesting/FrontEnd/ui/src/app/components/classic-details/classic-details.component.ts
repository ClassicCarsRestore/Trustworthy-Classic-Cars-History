import { Component, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import {Classic, Restoration} from '../../classic';
import { Router, NavigationExtras, ActivatedRoute } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import {Token} from '../../tokens';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UserLevel} from '../../user';
import { ClipboardService } from 'ngx-clipboard';
import { DatePipe } from '@angular/common';

@Component({
  selector: 'app-classic-details',
  templateUrl: './classic-details.component.html',
  styleUrls: ['./classic-details.component.css']
})
export class ClassicDetailsComponent {
  public access?: string;
  public classic?: Classic;
  public priorOwners?: string[];
  public restorations?: Restoration[];
  public showList: boolean[] = [];
  public certifications?: string[];
  public documents: Map<string, string> = new Map<string, string>();
  public documentsArray: Array<[string, string]> = [];
  public showSpinnerCertify = false;

  constructor(private http: HttpClient, private router: Router, private authService: AuthService, private activatedRoute: ActivatedRoute, private modalService: NgbModal, @Inject('BASE_URL') private baseUrl: string, private clipboardService: ClipboardService, private datePipe: DatePipe) {
    let chassisNo: string = "";

    // get id from url
    this.activatedRoute.paramMap.subscribe(params => {
      let tmpChassisNo: string | null = params.get('chassisNo');
      if (tmpChassisNo != null)
        chassisNo = tmpChassisNo;
    });

    http.get<UserLevel>(baseUrl+'api/Access/CheckLevel/'+chassisNo, Token.getHeader()).subscribe(result => {
      this.access = result.type;
    }, error => {
      console.log("An error occurred:", error);
    });

    http.get<Classic>(baseUrl + 'api/Classics/Get/' + chassisNo, Token.getHeader()).subscribe(result => {
      this.classic = result;
      this.priorOwners = result.priorOwners;
      this.restorations = result.restorations;
      this.certifications = result.certifications;
      this.documents = new Map<string, string>(Object.entries(result.documents));
      this.documentsArray = Array.from(this.documents.entries());
      console.log(this.documents);
    }, error => {
      if (error && error.error && (error.error.message === "404 - The classic "+chassisNo+" was Not Found" || error.error.message === "403 Forbidden - You are not authorized")) {
        alert(error.error.message);
        this.router.navigate(['/classics']);
      }
      console.log("An error occurred:", error);
      console.log("Error status:", error.status);
      console.log("Error message:", error.error.message);
      console.log("Error success:", error.error.success); 
    });
  }

  toggleList(index: number): void {
    this.showList[index] = !this.showList[index];
  }
  
  editClassic(classic: Classic) {
    let navigationExtras: NavigationExtras = {
      state: {
        classic: classic
      }
    };
    this.router.navigate(['/classics/details', classic.chassisNo, 'edit'], navigationExtras);
  }

  updateOwner(classic: Classic) {
    let navigationExtras: NavigationExtras = {
      state: {
        classic: classic
      }
    };
    this.router.navigate(['/classics/access/'+ classic.chassisNo], navigationExtras);
  }

  createStep(classic: Classic) {
    let navigationExtras: NavigationExtras = {
      state: {
        classic: classic
      }
    };
    this.router.navigate(['/restorations/create/'+ classic.chassisNo], navigationExtras);
  }

  copyLinkToClipboard() {
    const currentUrl = window.location.href;
    this.clipboardService.copyFromContent(currentUrl);
  }

  goToStep(chassisNo: string, stepId: string): void {
    this.router.navigate(['/restoration/details/'+chassisNo+'/'+ stepId]);
  }

  certifyClick(chassisNo: string) {
    this.showSpinnerCertify = true;
    this.http.put(this.baseUrl+ 'api/Classics/Certify/' + chassisNo,'', Token.getHeader()).subscribe(result => {
      this.navigateToDetails(chassisNo);
    }, error => {
      if (error && error.error && (error.error.message === "404 - The classic "+chassisNo+" was Not Found" || error.error.message === "403 Forbidden - You are not authorized")) {
        alert(error.error.message);
        this.router.navigate(['/classics/details/'+chassisNo])
        .then(() => {
          window.location.reload();
        });
      }
      console.log("An error occurred:", error); // This block is executed on error
    });
  }
  
  navigateToDetails(chassisNo: string) {
    const navigationExtras: NavigationExtras = {
      skipLocationChange: true
    };
  
    this.router.navigate(['/classics/details/', chassisNo], navigationExtras).then(() => {
      window.location.reload();
    });
  }

  addDocument(classic: Classic) {
    let navigationExtras: NavigationExtras = {
      state: {
        classic: classic
      }
    };
    this.router.navigate(['/classics/document/'+ classic.chassisNo], navigationExtras);
  }

  getHistory(classic: Classic) {
    let navigationExtras: NavigationExtras = {
      state: {
        classic: classic
      }
    };
    this.router.navigate(['/classics/history/'+classic.chassisNo], navigationExtras);
  }

  getAccessHistory(classic: Classic) {
    let navigationExtras: NavigationExtras = {
      state: {
        classic: classic
      }
    };
    this.router.navigate(['/classics/access/history/'+classic.chassisNo], navigationExtras);
  }

  formatDate(date: string): string {
    return this.datePipe.transform(date, 'dd-MM-y, HH:mm') || '';
  }

  getInfoByEmail(email: string) {
    this.router.navigate(['/users/'+email]);
  }

  handleImageError(event: Event) {
    const imgElement = event.target as HTMLImageElement;
    imgElement.src = 'assets/img/doc2.webp';
  }
}
