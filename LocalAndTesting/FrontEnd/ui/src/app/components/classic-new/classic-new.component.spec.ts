import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ClassicNewComponent } from './classic-new.component';

describe('ClassicNewComponent', () => {
  let component: ClassicNewComponent;
  let fixture: ComponentFixture<ClassicNewComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [ClassicNewComponent]
    });
    fixture = TestBed.createComponent(ClassicNewComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
