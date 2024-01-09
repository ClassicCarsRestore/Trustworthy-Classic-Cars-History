import { ComponentFixture, TestBed } from '@angular/core/testing';

import { RestorationCreateComponent } from './restoration-create.component';

describe('RestorationCreateComponent', () => {
  let component: RestorationCreateComponent;
  let fixture: ComponentFixture<RestorationCreateComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [RestorationCreateComponent]
    });
    fixture = TestBed.createComponent(RestorationCreateComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
