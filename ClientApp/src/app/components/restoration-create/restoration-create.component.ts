import { Component, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router, ActivatedRoute } from '@angular/router';
import {Token} from '../../tokens';
import { UserLevel} from '../../user';
import {Classic} from '../../classic';

@Component({
  selector: 'app-restoration-create',
  templateUrl: './restoration-create.component.html',
  styleUrls: ['./restoration-create.component.css']
})

export class RestorationCreateComponent {
  public classic?: Classic;
  public selectedFiles: File[] = [];

  public chassisNo?: string;
  public access?: string;
  public showSpinner = false;

  constructor(private http: HttpClient, private router: Router, private activatedRoute: ActivatedRoute, @Inject('BASE_URL') private baseUrl: string,) {
    this.baseUrl = baseUrl;

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
      if(result.type === 'owner' || result.type === 'modifier') {
        this.access = result.type;
      }
      else {
        alert('Not authorized.')
        this.router.navigate(['/classics/details/'+this.chassisNo]);
      }
    }, error => {
      this.router.navigate(['/classics/details/'+this.chassisNo]);
    });

  }

  onFileSelected(event: any): void {
    this.selectedFiles = event.target.files;
  }

  onSubmit() {
    let hasErrors: boolean = false;
    let title = (<HTMLInputElement>document.getElementById("title")).value;
    if (title.trim() == "") {
      hasErrors = true;
      document.getElementById("titleEmpty")!.innerHTML = "Restoration's title can't be empty";
    } else
      document.getElementById("titleEmpty")!.innerHTML = "";

    let description = (<HTMLInputElement>document.getElementById("description")).value;
    if (description.trim() == "") {
      hasErrors = true;
      document.getElementById("descriptionEmpty")!.innerHTML = "Restoration's description can't be empty";
    } else
      document.getElementById("descriptionEmpty")!.innerHTML = "";

    if(!hasErrors) {
      let confirmed = confirm("This operation may take a while (specially if adding images/documents). Do you want to proceed?");
      if(confirmed) {
        this.showSpinner = true;
        const formData = new FormData();
        formData.append('title', title);
        formData.append('description', description);
    
        for (let file of this.selectedFiles) {
          formData.append('file', file);
        }
    
        console.log(formData);
    
        this.http.post(this.baseUrl + 'api/Restorations/Create/'+this.chassisNo, formData, Token.getHeader())
          .subscribe(
            response => {
              console.log('Post request successful', response);
              this.router.navigate(['/classics/details/'+this.chassisNo]);
            },
            error => {
              if (error && error.error && (error.error.message === "404 - The classic "+this.chassisNo+" was Not Found" || error.error.message === "403 Forbidden - You are not authorized")) {
                alert("An error has occured with the creation of the procedure. Please refresh the page and try again.")
                console.error('Error submitting form', error);
              }
          });
      }
    }

  }

}
