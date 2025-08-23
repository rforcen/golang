#include <GLFW/glfw3.h>
#include <math.h>

#include "render.h"

void perspective_gl(float fov_y, float aspect, float z_near, float z_far) {
  float f_h = tanf(fov_y / 360.0f * M_PI) * z_near;
  float f_w = f_h * aspect;
  glFrustum(-f_w, f_w, -f_h, f_h, z_near, z_far);
}
void setGeo(int mousex, int mousey, float zoom, int w, int h) {

  glClear(GL_COLOR_BUFFER_BIT | GL_DEPTH_BUFFER_BIT);
  glClearColor(0.7f, 0.7f, 0.7f, 1.0f); // light bg

  glViewport(0, 0, w, h);
  glMatrixMode(GL_PROJECTION);
  glLoadIdentity();
  perspective_gl(45.0f, (float)w / (float)h, 0.1f, 100.0f);
  glMatrixMode(GL_MODELVIEW);
  glLoadIdentity();

  glTranslatef(0.0f, 0.0f, zoom);
  glRotatef(mousex, 0.0f, 1.0f, 0.0f);
  glRotatef(mousey, 1.0f, 0.0f, 0.0f);
}
void setPrefs() {
  glEnable(GL_LINE_SMOOTH);

  glHint(GL_LINE_SMOOTH_HINT, GL_NICEST);
  glHint(GL_POLYGON_SMOOTH_HINT, GL_NICEST);

  glClearDepth(1.0);       // Set background depth to farthest
  glEnable(GL_DEPTH_TEST); // Enable depth testing for z-culling
  glDepthFunc(GL_LEQUAL);  // Set the type of depth-test
  glShadeModel(GL_SMOOTH); // Enable smooth shading
  glHint(GL_PERSPECTIVE_CORRECTION_HINT,
         GL_NICEST); //  Nice perspective corrections
}

void sceneInit() { // works nice for golden solid colors (requires normals)
  float lmodel_ambient[] = {0, 0, 0, 0};
  float lmodel_twoside[] = {GL_FALSE};
  float light0_ambient[] = {0.1f, 0.1f, 0.1f, 1.0f};
  float light0_diffuse[] = {1.0f, 1.0f, 1.0f, 0.0f};
  float light0_position[] = {1, 0.5, 1, 0};
  float light1_position[] = {-1, 0.5, -1, 0};
  float light0_specular[] = {1, 1, 1, 0};
  float bevel_mat_ambient[] = {0, 0, 0, 1};
  float bevel_mat_shininess[] = {40};
  float bevel_mat_specular[] = {1, 1, 1, 0};
  float bevel_mat_diffuse[] = {1, 0, 0, 0};

  //  glClearColor(float(color.redF()), float(color.greenF()),
  //  float(color.blueF()),               1);

  glLightfv(GL_LIGHT0, GL_AMBIENT, light0_ambient);
  glLightfv(GL_LIGHT0, GL_DIFFUSE, light0_diffuse);
  glLightfv(GL_LIGHT0, GL_SPECULAR, light0_specular);
  glLightfv(GL_LIGHT0, GL_POSITION, light0_position);
  glEnable(GL_LIGHT0);

  glLightfv(GL_LIGHT1, GL_AMBIENT, light0_ambient);
  glLightfv(GL_LIGHT1, GL_DIFFUSE, light0_diffuse);
  glLightfv(GL_LIGHT1, GL_SPECULAR, light0_specular);
  glLightfv(GL_LIGHT1, GL_POSITION, light1_position);
  glEnable(GL_LIGHT1);

  glLightModelfv(GL_LIGHT_MODEL_TWO_SIDE, lmodel_twoside);
  glLightModelfv(GL_LIGHT_MODEL_AMBIENT, lmodel_ambient);
  glEnable(GL_LIGHTING);

  glMaterialfv(GL_FRONT, GL_AMBIENT, bevel_mat_ambient);
  glMaterialfv(GL_FRONT, GL_SHININESS, bevel_mat_shininess);
  glMaterialfv(GL_FRONT, GL_SPECULAR, bevel_mat_specular);
  glMaterialfv(GL_FRONT, GL_DIFFUSE, bevel_mat_diffuse);

  glEnable(GL_COLOR_MATERIAL);
  glShadeModel(GL_SMOOTH);

  glEnable(GL_LINE_SMOOTH);

  glHint(GL_LINE_SMOOTH_HINT, GL_NICEST);
  glHint(GL_POLYGON_SMOOTH_HINT, GL_NICEST);

  glClearDepth(1.0);       // Set background depth to farthest
  glEnable(GL_DEPTH_TEST); // Enable depth testing for z-culling
  glDepthFunc(GL_LEQUAL);  // Set the type of depth-test
  glShadeModel(GL_SMOOTH); // Enable smooth shading
  glHint(GL_PERSPECTIVE_CORRECTION_HINT,
         GL_NICEST); //  Nice perspective corrections
}

// This function needs to be called inside your main rendering loop.
void drawMesh(CCoord *vertexes, int vertexes_count, int *faces, int faces_count,
              CCoord *colors, int colors_count) {
  glBegin(GL_POLYGON);
  for (int i = 0; i < faces_count; i++) {
    glColor3f(colors[i].x, colors[i].y, colors[i].z);
    glVertex3f(vertexes[faces[i]].x, vertexes[faces[i]].y,
               vertexes[faces[i]].z);
  }
  glEnd();
}