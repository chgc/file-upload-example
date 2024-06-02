import { JsonPipe } from '@angular/common';
import { HttpClient, HttpEventType } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { bufferCount, concatMap, forkJoin, from, tap, toArray } from 'rxjs';

interface UploadStatus {
  fileId: string;
  upload: number;
  total: number;
}
@Component({
  selector: 'app-root',
  standalone: true,
  imports: [ReactiveFormsModule, JsonPipe],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css',
})
export class AppComponent {
  http = inject(HttpClient);

  title = 'frontend';

  uploadProgress = signal<UploadStatus[]>([]);

  formData = new FormGroup({
    files: new FormControl<FileList | null>(null),
  });

  onFileChange(event: any) {
    console.log(event.target.files);
    if (event.target.files.length > 0) {
      const file = event.target.files;
      this.formData.patchValue({
        files: file,
      });
    }
  }

  upload() {
    const url = 'http://localhost:8080/upload';
    const files = this.formData.controls.files.value;

    if (files) {
      const chunkPromises = Array.from(files)
        .map((file) => {
          const chunkSize = 5 * 1024 * 1024; // 5MB
          const chunks = Math.ceil(file.size / chunkSize);
          const chunkPromises = [];
          this.uploadProgress.update((status) => {
            let s = status.find((x) => x.fileId === file.name);
            if (s == null) {
              return [
                ...status,
                { fileId: file.name, total: chunks, upload: 0 },
              ];
            }
            return status;
          });
          for (let i = 0; i < chunks; i++) {
            const start = i * chunkSize;
            const end = Math.min(file.size, start + chunkSize);
            const chunk = file.slice(start, end);

            const formData = new FormData();
            formData.append('file_id', file.name);
            formData.append('chunk', chunk, file.name);
            formData.append('chunkIndex', String(i + 1));

            const request = this.http
              .request('POST', url, {
                body: formData,
                observe: 'events',
                reportProgress: true,
              })
              .pipe(
                tap((event) => {
                  if (event.type === HttpEventType.Response) {
                    this.uploadProgress.update((status) => {
                      let s = status.find((x) => x.fileId === file.name);
                      if (s != null) {
                        s.upload += 1;
                      }
                      return [...status];
                    });
                  }
                })
              );
            chunkPromises.push(request);
          }
          return chunkPromises;
        })
        .flatMap((x) => x);

      from(chunkPromises)
        .pipe(
          bufferCount(10),
          concatMap((reqs) => forkJoin(reqs)),
          toArray(),
          concatMap(() =>
            forkJoin(
              this.uploadProgress().map((file) =>
                this.http.post('http://localhost:8080/upload/complete', {
                  fileId: file.fileId,
                  totalChunks: file.total,
                })
              )
            )
          )
        )
        .subscribe({
          next: (results) => {
            console.log('All chunks uploaded successfully', results);
          },
          error: (error) => {
            console.error('Error uploading chunks', error);
          },
        });
    }
  }
}
