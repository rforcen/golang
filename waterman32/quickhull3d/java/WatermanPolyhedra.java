
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;

public // WatermanPolyhedra.
class WatermanPolyhedra {
    int ntc;
    Coords coords;
    int rad = 0;

    WatermanPolyhedra(int rad) { // generate coords & faces of the convex hull
        coords = null;
        this.rad = rad;
        double[] dcoords = WatermanData(rad); // Math.sqrt((double) rad)); // generate double coords from SQRT(rad)!!
        ConvexHull(dcoords);
        coords.normalise();
        coords.genFloatCoords(); // -> coords.fcoords
    }

    double[] WatermanData(double radius) { // retrun total number of vertex (3*ncoords)
        int counter = 0;
        double x, y, z, a, b, c, xra, xrb, yra, yrb, zra, zrb, R, Ry, s, radius2;
        if (radius > 350)
            return null; // max radius 20*20=400

        double[] coords = null; // max rad=500

        a = b = c = 0; // center

        for (int pass = 0; pass < 2; pass++) { // 1st pass=calc counter, second pass =assign coords

            if (pass == 1)
                coords = new double[counter];

            counter = 0;
            s = radius;
            radius2 = radius * radius;
            xra = Math.ceil(a - s);
            xrb = Math.floor(a + s);

            for (x = xra; x <= xrb; x++) {
                R = radius2 - (x - a) * (x - a);
                if (R < 0)
                    continue;
                s = Math.sqrt(R);
                yra = Math.ceil(b - s);
                yrb = Math.floor(b + s);
                for (y = yra; y <= yrb; y++) {
                    Ry = R - (y - b) * (y - b);
                    if (Ry < 0)
                        continue; // case Ry < 0
                    if (Ry == 0 && c == Math.floor(c)) { // case Ry=0
                        if (((x + y + c) % 2) != 0)
                            continue;
                        else {
                            zra = c;
                            zrb = c;
                        }
                    } else { // case Ry > 0
                        s = Math.sqrt(Ry);
                        zra = Math.ceil(c - s);
                        zrb = Math.floor(c + s);
                        if (((x + y) % 2) == 0) {// (x+y)mod2=0
                            if ((zra % 2) != 0) {
                                if (zra <= c)
                                    zra = zra + 1;
                                else
                                    zra = zra - 1;
                            }
                        } else { // (x+y) mod 2 <> 0
                            if ((zra % 2) == 0) {
                                if (zra <= c)
                                    zra = zra + 1;
                                else
                                    zra = zra - 1;
                            }
                        }
                    }

                    for (z = zra; z <= zrb; z += 2) { // save vertex x,y,z
                        if (pass == 1) {
                            coords[counter++] = x;
                            coords[counter++] = y;
                            coords[counter++] = z;
                        } else
                            counter += 3;
                    }

                }
            }
        }

        ntc = counter; // set total number of coords
        double[] crd = new double[ntc]; // create a copy of used coords
        for (int i = 0; i < ntc; i++)
            crd[i] = coords[i];
        coords = null; // don't need any more
        return crd;
    }

    void ConvexHull(double[] coords) {
        QuickHull3D hull = new QuickHull3D();
        hull.build(coords); // generate hull from points
        Point3d[] vertices = hull.getVertices(); // get vertex
        int[][] faceIndices = hull.getFaces(); // get faces

        int nc = vertices.length, nf = faceIndices.length;
        this.coords = new Coords(nc, nf); // new coords instance

        // this.coords.coords[]=vertices[]
        for (int i = 0; i < nc; i++) {
            Point3d pnt = vertices[i];
            Coords.Point3d cp = this.coords.coords[i];
            cp.x = (float) pnt.x;
            cp.y = (float) pnt.y;
            cp.z = (float) pnt.z;
        }

        // this.coords.faces[]=(n,faceIndeces[])
        for (int i = 0; i < nf; i++) {
            int n = faceIndices[i].length;
            this.coords.face[i].add(n, faceIndices[i]);
        }

    }

    String print() {
        // format
        // root, n.coords, n.faces
        // coords
        // faces(coord index set)
        String s = "";

        s += String.format("%d,%d,%d\n", rad, coords.nc, coords.nf);
        for (int i = 0; i < coords.nc; i++) {
            Coords.Point3d p = coords.coords[i];
            s += String.format("%f,%f,%f%c", p.x, p.y, p.z, (i < coords.nc - 1) ? ',' : '\n');
        }

        for (int f = 0; f < coords.nf; f++) {
            Coords.Face face = coords.face[f];
            for (int i = 0; i < face.n; i++)
                s += String.format("%d%c", face.coords[i], (i < face.n - 1) ? ',' : '\n');
        }
        return s;
    }

    // wc={2:[[ [0,0,0],[1,1,1] ] ,[[0,1,2],[3,4,5,6]],
    // 3:[[[0,0,0],[1,1,1]],[0,1,2]]}
    String toPython() {
        String s = "";

        s += String.format("%d:[[", rad);
        for (int i = 0; i < coords.nc; i++) {
            Coords.Point3d p = coords.coords[i];
            s += String.format("[%f,%f,%f]%c", p.x, p.y, p.z, (i < coords.nc - 1) ? ',' : ']');
        }
        s += ", [";

        for (int f = 0; f < coords.nf; f++) {
            Coords.Face face = coords.face[f];
            s += "[";
            for (int i = 0; i < face.n; i++)
                s += String.format("%d%c", face.coords[i], (i < face.n - 1) ? ',' : ']');
            s += (f < coords.nf - 1) ? ", " : "]]";
        }
        return s;
    }

    // rad nCoords coordsx3....
    // nf
    String toText() {
        String s = "";

        s += String.format("%d %d ", rad, coords.nc);
        for (int i = 0; i < coords.nc; i++) {
            Coords.Point3d p = coords.coords[i];
            s += String.format("%.3f %.3f %.3f ", p.x, p.y, p.z);
        }
        s += String.format("%d ", coords.nf);

        for (int f = 0; f < coords.nf; f++) {
            Coords.Face face = coords.face[f];
            s += String.format("%d ", face.n);
            for (int i = 0; i < face.n; i++)
                s += String.format("%d ", face.coords[i]);
            s += (f < coords.nf) ? "" : "\n";
        }
        return s;
    }

    void save(Path fName, byte[] s) throws IOException {
        Path write;
        write = Files.write(fName, s);
    }
}
