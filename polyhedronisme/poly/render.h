#pragma once

#include <GLFW/glfw3.h>
#include <math.h>

typedef struct {
  float x, y, z;
} CCoord;


extern void perspective_gl(float fov_y, float aspect, float z_near, float z_far);
extern void setGeo(int mousex, int mousey, float zoom, int w, int h);
extern void sceneInit();
extern void drawMesh(CCoord *vertexes, int vertexes_count, int *faces, int faces_count, CCoord*colors, int colors_count);
extern void setPrefs();
