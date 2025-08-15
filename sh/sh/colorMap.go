package sh

const N_COLOR_MAPS = 25

// ColorMap
func ColorMap(v_ float32, vmin_ float32, vmax_ float32, type_ int) Coord {
	v := v_
	vmin := vmin_
	vmax := vmax_

	dv := float32(0.0)
	//vmid := float32(0.0)

	c := NewCoord(1.0, 1.0, 1.0)

	c1 := NewCoord(0.0, 0.0, 0.0)
	c2 := NewCoord(0.0, 0.0, 0.0)
	c3 := NewCoord(0.0, 0.0, 0.0)

	if vmax < vmin {
		dv = vmin
		vmin = vmax
		vmax = dv
	}
	if vmax-vmin < 1.0e-6 {
		vmin -= 1.0
		vmax += 1.0
	}
	if v < vmin {
		v = vmin
	}
	if v > vmax {
		v = vmax
	}
	dv = vmax - vmin

	switch type_ {
	case 1:
		if v < (vmin + 0.25*dv) {
			c = NewCoord(0.0, 4.0*(v-vmin)/dv, 1.0)
		} else if v < (vmin + 0.5*dv) {
			c = NewCoord(0.0, 1.0, c.Z)
			c.Z = 1.0 + 4.0*(vmin+0.25*dv-v)/dv
		} else if v < (vmin + 0.75*dv) {
			c = NewCoord(4.0*(v-vmin-0.25*dv)/dv, 1.0, 1.0+4.0*(vmin+0.25*dv-v)/dv)
			c.X = 4.0 * (v - vmin - 0.5*dv) / dv
			c.Y = 1.0
			c.Z = 0.0
		} else {
			c = NewCoord(1.0, 1.0, 4.0*(v-vmin-0.75*dv)/dv)
		}
	case 2:
		c = NewCoord((v-vmin)/dv, 0.0, (vmax-v)/dv)
	case 3:
		val := (v - vmin) / dv
		c = NewCoord(val, val, val)
	case 4:
		if v < (vmin + dv/6.0) {
			c = NewCoord(1.0, 6.0*(v-vmin)/dv, 0.0)
		} else if v < (vmin + 2.0*dv/6.0) {
			c = NewCoord(1.0+6.0*(vmin+dv/6.0-v)/dv, 1.0, 0.0)
		} else if v < (vmin + 3.0*dv/6.0) {
			c = NewCoord(0.0, 1.0, 6.0*(v-vmin-2.0*dv/6.0)/dv)
		} else if v < (vmin + 4.0*dv/6.0) {
			c = NewCoord(1.0, 1.0+4.0*(vmin+0.5*dv-v)/dv, 0.0)
		} else if v < (vmin + 5.0*dv/6.0) {
			c.X = 6.0 * (v - vmin - 4.0*dv/6.0) / dv
			c.Y = 0.0
		} else {
			c.X = 1.0
			c.Y = 0.0
			c.Z = 1.0 + 6.0*(vmin+5.0*dv/6.0-v)/dv
		}
	case 5:
		c = NewCoord((v-vmin)/dv, 1.0, 0.0)
	case 6:
		val := (v - vmin) / (vmax - vmin)
		c = NewCoord(val, (vmax-v)/(vmax-vmin), val)
	case 7:
		if v < (vmin + 0.25*dv) {
			val := 4.0 * (v - vmin) / dv
			c = NewCoord(0.0, val, 1.0-val)
		} else if v < (vmin + 0.5*dv) {
			val := 4.0 * (v - vmin - 0.25*dv) / dv
			c = NewCoord(val, 1.0-val, 0.0)
		} else if v < (vmin + 0.75*dv) {
			val := 4.0 * (v - vmin - 0.5*dv) / dv
			c = NewCoord(1.0-val, val, 0.0)
		} else {
			c = NewCoord(1.0, 4.0*(v-vmin-0.75*dv)/dv, 0.0)
		}
	case 8:
		if v < (vmin + 0.5*dv) {
			val := 2.0 * (v - vmin) / dv
			c = NewCoord(val, val, val)
		} else {
			val := 1.0 - 2.0*(v-vmin-0.5*dv)/dv
			c = NewCoord(val, val, val)
		}
	case 9:
		if v < (vmin + dv/3.0) {
			val := 3.0 * (v - vmin) / dv
			c = NewCoord(1.0-val, 0.0, val)
		} else if v < (vmin + 2.0*dv/3.0) {
			c = NewCoord(0.0, 3.0*(v-vmin-dv/3.0)/dv, 1.0)
		} else {
			val := 3.0 * (v - vmin - 2.0*dv/3.0) / dv
			c = NewCoord(val, 1.0-val, 1.0)
		}
	case 10:
		if v < (vmin + 0.2*dv) {
			c = NewCoord(0.0, 5.0*(v-vmin)/dv, 1.0)
		} else if v < (vmin + 0.4*dv) {
			c = NewCoord(0.0, 1.0, 1.0+5.0*(vmin+0.2*dv-v)/dv)
		} else if v < (vmin + 0.6*dv) {
			c = NewCoord(5.0*(v-vmin-0.4*dv)/dv, 1.0, 0.0)
		} else if v < (vmin + 0.8*dv) {
			c = NewCoord(1.0, 1.0-5.0*(v-vmin-0.6*dv)/dv, 0.0)
		} else {
			val := 5.0 * (v - vmin - 0.8*dv) / dv
			c = NewCoord(1.0, val, val)
		}
	case 11:
		c1 = NewCoord(200.0/255, 60.0/255, 0.0/255)
		c2 = NewCoord(250.0/255, 160.0/255, 110.0/255)
		t := (v - vmin) / dv
		c = *c1.MulScalar(1 - t).Add(c2.MulScalar(t))
	case 12:
		c1 = NewCoord(55.0/255, 55.0/255, 45.0/255)
		c2 = NewCoord(235.0/255, 90.0/255, 30.0/255)
		c3 = NewCoord(250.0/255, 160.0/255, 110.0/255)
		ratio := float32(0.4)
		vmid := vmin + ratio*dv
		if v < vmid {
			t := (v - vmin) / (ratio * dv)
			c = *c1.MulScalar(1 - t).Add(c2.MulScalar(t))
		} else {
			t := (v - vmid) / ((1.0 - ratio) * dv)
			c = *c2.MulScalar(1 - t).Add(c3.MulScalar(t))
		}
	case 13:
		c1 = NewCoord(0.0/255, 255.0/255, 0.0/255)
		c2 = NewCoord(255.0/255, 150.0/255, 0.0/255)
		c3 = NewCoord(255.0/255, 250.0/255, 240.0/255)
		ratio := float32(0.3)
		vmid := vmin + ratio*dv
		if v < vmid {
			c.X = (c2.X-c1.X)*(v-vmin)/(ratio*dv) + c1.X
			c.Y = (c2.Y-c1.Y)*(v-vmin)/(ratio*dv) + c1.Y
			c.Z = (c2.Z-c1.Z)*(v-vmin)/(ratio*dv) + c1.Z
		} else {
			c.X = (c3.X-c2.X)*(v-vmid)/((1.0-ratio)*dv) + c2.X
			c.Y = (c3.Y-c2.Y)*(v-vmid)/((1.0-ratio)*dv) + c2.Y
			c.Z = (c3.Z-c2.Z)*(v-vmid)/((1.0-ratio)*dv) + c2.Z
		}
	case 14:
		c = NewCoord(1.0, 1.0-(v-vmin)/dv, 0.0)
	case 15:
		if v < (vmin + 0.25*dv) {
			c = NewCoord(0.0, 4.0*(v-vmin)/dv, 1.0)
		} else if v < (vmin + 0.5*dv) {
			c = NewCoord(0.0, 1.0, 1.0-4.0*(v-vmin-0.25*dv)/dv)
		} else if v < (vmin + 0.75*dv) {
			c = NewCoord(4.0*(v-vmin-0.5*dv)/dv, 1.0, 0.0)
		} else {
			c = NewCoord(1.0, 1.0, 4.0*(v-vmin-0.75*dv)/dv)
		}
	case 16:
		if v < (vmin + 0.5*dv) {
			c = NewCoord(0.0, 2.0*(v-vmin)/dv, 1.0-2.0*(v-vmin)/dv)
		} else {
			c = NewCoord(2.0*(v-vmin-0.5*dv)/dv, 1.0, 2.0*(v-vmin-0.5*dv)/dv)
		}
	case 17:
		if v < (vmin + 0.5*dv) {
			c = NewCoord(1.0, 1.0-2.0*(v-vmin)/dv, 2.0*(v-vmin)/dv)
		} else {
			c = NewCoord(1.0-2.0*(v-vmin-0.5*dv)/dv, 2.0*(v-vmin-0.5*dv)/dv, 1.0)
		}
	case 18:
		c = NewCoord(0.0, (v-vmin)/(vmax-vmin), 1.0)
	case 19:
		c = NewCoord((v-vmin)/(vmax-vmin), (v-vmin)/(vmax-vmin), 1.0)
	case 20:
		c1 = NewCoord(0.0/255, 160.0/255, 0.0/255)
		c2 = NewCoord(180.0/255, 220.0/255, 0.0/255)
		c3 = NewCoord(250.0/255, 220.0/255, 170.0/255)
		ratio := float32(0.3)
		vmid := vmin + ratio*dv
		if v < vmid {
			c = NewCoord((c2.X-c1.X)*(v-vmin)/(ratio*dv)+c1.X, (c2.Y-c1.Y)*(v-vmin)/(ratio*dv)+c1.Y, (c2.Z-c1.Z)*(v-vmin)/(ratio*dv)+c1.Z)
		} else {
			c = NewCoord((c3.X-c2.X)*(v-vmid)/((1.0-ratio)*dv)+c2.X, (c3.Y-c2.Y)*(v-vmid)/((1.0-ratio)*dv)+c2.Y, (c3.Z-c2.Z)*(v-vmid)/((1.0-ratio)*dv)+c2.Z)
		}
	case 21:
		c1 = NewCoord(255.0/255, 255.0/255, 200.0/255)
		c2 = NewCoord(150.0/255, 150.0/255, 255.0/255)
		c = NewCoord((c2.X-c1.X)*(v-vmin)/dv+c1.X, (c2.Y-c1.Y)*(v-vmin)/dv+c1.Y, (c2.Z-c1.Z)*(v-vmin)/dv+c1.Z)
	case 22:
		c = NewCoord(1.0-(v-vmin)/dv, 1.0-(v-vmin)/dv, (v-vmin)/dv)
	case 23:
		if v < (vmin + 0.5*dv) {
			c = NewCoord(1.0, 2.0*(v-vmin)/dv, 2.0*(v-vmin)/dv)
		} else {
			c = NewCoord(1.0-2.0*(v-vmin-0.5*dv)/dv, 1.0-2.0*(v-vmin-0.5*dv)/dv, 1.0)
		}
	case 24:
		if v < (vmin + 0.5*dv) {
			c = NewCoord(2.0*(v-vmin)/dv, 2.0*(v-vmin)/dv, 1.0-2.0*(v-vmin)/dv)
		} else {
			c = NewCoord(1.0-2.0*(v-vmin-0.5*dv)/dv, 1.0-2.0*(v-vmin-0.5*dv)/dv, 0.0)
		}
	case 25:
		if v < (vmin + dv/3) {
			c = NewCoord(0.0, 3.0*(v-vmin)/dv, 1.0)
		} else if v < (vmin + 2*dv/3) {
			c = NewCoord(3.0*(v-vmin-dv/3)/dv, 1.0-3.0*(v-vmin-dv/3)/dv, 1.0)
		} else {
			c = NewCoord(1.0, 0.0, 1.0-3.0*(v-vmin-2*dv/3)/dv)
		}
	}
	return c
}
