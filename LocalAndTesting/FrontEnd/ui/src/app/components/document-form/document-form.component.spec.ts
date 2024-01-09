import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DocumentFormComponent } from './document-form.component';

describe('DocumentFormComponent', () => {
  let component: DocumentFormComponent;
  let fixture: ComponentFixture<DocumentFormComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [DocumentFormComponent]
    });
    fixture = TestBed.createComponent(DocumentFormComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
