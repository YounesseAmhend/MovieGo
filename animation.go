package moviego

type Animatable interface {
	HasAnimation() bool
	GetPositionAnim() *PositionAnimParams
	GetRotationAnim() *RotationAnimParams
	GetScaleAnim() *ScaleAnimParams
	GetStartTime() float64
	GetDuration() float64
}

