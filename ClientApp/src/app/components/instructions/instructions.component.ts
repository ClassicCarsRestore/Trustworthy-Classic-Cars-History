import { Component, ViewChild, ElementRef, AfterViewInit } from '@angular/core';

@Component({
  selector: 'app-instructions',
  templateUrl: './instructions.component.html',
  styleUrls: ['./instructions.component.css']
})
export class InstructionsComponent implements AfterViewInit {
  @ViewChild('loginInfo') loginInfo!: ElementRef;
  @ViewChild('registerInfo') registerInfo!: ElementRef;
  @ViewChild('classicsInfo') classicsInfo!: ElementRef;
  @ViewChild('vehicleCardInfo') vehicleCardInfo!: ElementRef;
  @ViewChild('editInfo') editInfo!: ElementRef;
  @ViewChild('historyInfo') historyInfo!: ElementRef;
  @ViewChild('accessInfo') accessInfo!: ElementRef;
  @ViewChild('accessHistoryInfo') accessHistoryInfo!: ElementRef;
  @ViewChild('newRestorationInfo') newRestorationInfo!: ElementRef;
  @ViewChild('updateRestorationInfo') updateRestorationInfo!: ElementRef;
  @ViewChild('addDocumentInfo') addDocumentInfo!: ElementRef;
  @ViewChild('createClassicInfo') createClassicInfo!: ElementRef;
  @ViewChild('restorationClassicsInfo') restorationClassicsInfo!: ElementRef;
  @ViewChild('certifierClassicsInfo') certifierClassicsInfo!: ElementRef;
  @ViewChild('addCertificationInfo') addCertificationInfo!: ElementRef;

  ngAfterViewInit() {
    const fragment = window.location.hash.substring(1);
    if (fragment === 'loginInfo') {
      this.loginInfo.nativeElement.scrollIntoView({ behavior: 'smooth' });
    } else if (fragment === 'registerInfo') {
      this.registerInfo.nativeElement.scrollIntoView({ behavior: 'smooth' });
    } else if (fragment === 'classicsInfo') {
      this.classicsInfo.nativeElement.scrollIntoView({ behavior: 'smooth' });
    } else if (fragment === 'vehicleCardInfo') {
      this.vehicleCardInfo.nativeElement.scrollIntoView({ behavior: 'smooth' });
    } else if (fragment === 'editInfo') {
      this.editInfo.nativeElement.scrollIntoView({ behavior: 'smooth' });
    } else if (fragment === 'historyInfo') {
      this.historyInfo.nativeElement.scrollIntoView({ behavior: 'smooth' });
    } else if (fragment === 'accessInfo') {
      this.accessInfo.nativeElement.scrollIntoView({ behavior: 'smooth' });
    } else if (fragment === 'accessHistoryInfo') {
      this.accessHistoryInfo.nativeElement.scrollIntoView({ behavior: 'smooth' });
    } else if (fragment === 'newRestorationInfo') {
      this.newRestorationInfo.nativeElement.scrollIntoView({ behavior: 'smooth' });
    } else if (fragment === 'updateRestorationInfo') {
      this.updateRestorationInfo.nativeElement.scrollIntoView({ behavior: 'smooth' });
    } else if (fragment === 'addDocumentInfo') {
      this.addDocumentInfo.nativeElement.scrollIntoView({ behavior: 'smooth' });
    } else if (fragment === 'createClassicInfo') {
      this.createClassicInfo.nativeElement.scrollIntoView({ behavior: 'smooth' });
    } else if (fragment === 'restorationClassicsInfo') {
      this.restorationClassicsInfo.nativeElement.scrollIntoView({ behavior: 'smooth' });
    } else if (fragment === 'certifierClassicsInfo') {
      this.certifierClassicsInfo.nativeElement.scrollIntoView({ behavior: 'smooth' });
    } else if (fragment === 'addCertificationInfo') {
      this.addCertificationInfo.nativeElement.scrollIntoView({ behavior: 'smooth' });
    } 
  }
}
