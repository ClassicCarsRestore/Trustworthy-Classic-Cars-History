import { Component, Inject} from '@angular/core';
import { HttpClient } from '@angular/common/http';
import {Token} from '../../tokens';
import { Router, ActivatedRoute } from '@angular/router';
import { AuthService } from '../../services/auth.service'
import {Classic} from '../../classic';

@Component({
  selector: 'app-document-form',
  templateUrl: './document-form.component.html',
  styleUrls: ['./document-form.component.css']
})
export class DocumentFormComponent {
  public classic?: Classic;
  public chassisNo?: string;
  public selectedFiles: File[] = [];
  public showSpinner = false;

  constructor(private authService: AuthService, private http: HttpClient, private router: Router, private activatedRoute: ActivatedRoute, @Inject('BASE_URL') private baseUrl: string,) {
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
  }

  onFileSelected(event: any): void {
    this.selectedFiles = event.target.files;
  }

  onSubmit() {
    let hasErrors: boolean = false;

    let title = (<HTMLInputElement>document.getElementById("title")).value;
    if (title.trim() == "") {
      hasErrors = true;
      document.getElementById("titleEmpty")!.innerHTML = "Document's title can't be empty";
    } else {
      document.getElementById("titleEmpty")!.innerHTML = "";
    }

    if(this.selectedFiles.length == 0) {
      hasErrors = true;
      document.getElementById("documentEmpty")!.innerHTML = "Please add a document";
    }
    else {
      document.getElementById("documentEmpty")!.innerHTML = "";
    }

    if(!hasErrors) {
      this.showSpinner = true;
      const formData = new FormData();
      formData.append('name', title);
      for (let file of this.selectedFiles) {
        formData.append('file', file);
      }
      this.http.put(this.baseUrl+ 'api/Classics/AddDocument/' + this.chassisNo, formData, Token.getHeader()).subscribe(result => {
        console.log(result);
        this.router.navigate(['/classics/details/'+ this.chassisNo]);
      }, error => {
        if (error && error.error && (error.error.message === "404 - The classic "+this.chassisNo+" was Not Found" || error.error.message === "403 Forbidden - You are not authorized")) {
          alert(error.error.message);
          this.router.navigate(['/classics/details/'+this.chassisNo]);
        }
        console.log("An error occurred:", error); // This block is executed on error
      });

    }

  }
}
