#pragma once
#include <GLFW/glfw3.h>
#include <math.h>

typedef struct {
  float x, y, z;
} CCoord;

typedef struct {
  CCoord coord;
  CCoord normal;
  CCoord color;
  CCoord uv;
} CLocation;


extern void perspective_gl(float fov_y, float aspect, float z_near, float z_far);
extern void setGeo(int mousex, int mousey, float zoom, int w, int h);
extern void sceneInit();
extern void drawMesh(CLocation *mesh, int res);
