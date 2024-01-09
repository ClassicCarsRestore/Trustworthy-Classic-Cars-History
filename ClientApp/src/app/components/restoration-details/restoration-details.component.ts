import { Component, OnInit, Inject } from '@angular/core';
import { AuthService } from '../../services/auth.service'
import { Observable } from 'rxjs';
import { HttpClient } from '@angular/common/http';
import { Router, ActivatedRoute } from '@angular/router';
import {Restoration} from '../../classic';
import {Token} from '../../tokens';

@Component({
  selector: 'app-restoration-details',
  templateUrl: './restoration-details.component.html',
  styleUrls: ['./restoration-details.component.css']
})
export class RestorationDetailsComponent implements OnInit{
  public loggedInOrg$: Observable<string>;
  public chassisNo?: string;
  public stepId?: string;
  public restoration?: Restoration;
  public showSpinner = false;
  public selectedFiles: File[] = [];
  public showSpinner2 = false;

  constructor(private authService: AuthService, private http: HttpClient, private router: Router, private activatedRoute: ActivatedRoute, @Inject('BASE_URL') private baseUrl: string,) { 
    this.loggedInOrg$ = this.authService.loggedInOrgObservable;

    // get id from url
    this.activatedRoute.paramMap.subscribe(params => {
      let tmpChassisNo: string | null = params.get('chassisNo');
      if (tmpChassisNo != null)
        this.chassisNo = tmpChassisNo;
    });

    this.activatedRoute.paramMap.subscribe(params => {
      let tmpStepId: string | null = params.get('stepId');
      if (tmpStepId != null)
        this.stepId = tmpStepId;
    });
    
    this.http.get<Restoration>(baseUrl + 'api/Restorations/Get/' + this.chassisNo +'/'+this.stepId, Token.getHeader()).subscribe(result => {
      this.restoration = result;
      console.log(this.restoration);
    }, error => {
      if (error && error.error && (error.error.message === "404 - The classic "+this.chassisNo+" was Not Found" || error.error.message === "403 Forbidden - You are not authorized" || error.error.message === "The step "+this.stepId +" was not found.")) {
        alert(error.error.message);
        this.router.navigate(['/classics/details/'+this.chassisNo]);
      }
      console.log("An error occurred:", error);
    });
  }
    
  //not used
  ngOnInit() {
    this.loggedInOrg$.subscribe(value => {
      console.log(value);
    });
  }

  onSubmitForm() {
    let hasErrors: boolean = false;

    let title = (<HTMLInputElement>document.getElementById("title")).value;
    if (title.trim() == "") {
      hasErrors = true;
      document.getElementById("titleEmpty")!.innerHTML = "Restoration's title can't be empty";
    } else {
      document.getElementById("titleEmpty")!.innerHTML = "";
    }

    let description = (<HTMLInputElement>document.getElementById("description")).value;
    if (description.trim() == "") {
      hasErrors = true;
      document.getElementById("descriptionEmpty")!.innerHTML = "Restoration's description can't be empty";
    } else {
      document.getElementById("descriptionEmpty")!.innerHTML = "";
    }

    if(!hasErrors) {
      let confirmed = confirm("This operation may take a while (specially if adding images/documents). Do you want to proceed?");
      if(confirmed) {
        this.showSpinner = true;
        const formData = new FormData();
        formData.append('stepId', String(this.stepId));
        formData.append('newTitle', title);
        formData.append('newDescription', description);
        if(this.selectedFiles.length > 0) {
          for (let file of this.selectedFiles) {
            formData.append('file', file);
          }
        }
        this.http.put(this.baseUrl+ 'api/Restorations/UpdateAndPhotos/' + this.chassisNo, formData, Token.getHeader()).subscribe(result => {
          console.log(result);
          this.router.navigate(['/classics/details/'+ this.chassisNo]);
        }, error => {
          if (error && error.error && (error.error.message === "404 - The classic "+this.chassisNo+" was Not Found" || error.error.message === "403 Forbidden - You are not authorized" || error.error.message === "The step "+this.stepId +" was not found." || error.error.message === "Not authorized to update step "+this.stepId +" since you are who made it originally.")) {
            alert(error.error.message);
            this.router.navigate(['/classics/details/'+this.chassisNo]);
          }
          console.log("An error occurred:", error); // This block is executed on error
          console.log("Error status:", error.status); // Access error status
          console.log("Error message:", error.error.message); // Access error message
          console.log("Error success:", error.error.success); // Access error success field
        });
      }
    }
  }

  onFileSelected(event: any): void {
    this.selectedFiles = event.target.files;
  }

}
