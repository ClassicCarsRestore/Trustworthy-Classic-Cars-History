import { ComponentFixture, TestBed } from '@angular/core/testing';

import { OwnerClassicsComponent } from './owner-classics.component';

describe('OwnerClassicsComponent', () => {
  let component: OwnerClassicsComponent;
  let fixture: ComponentFixture<OwnerClassicsComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [OwnerClassicsComponent]
    });
    fixture = TestBed.createComponent(OwnerClassicsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
