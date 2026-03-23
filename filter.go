package moviego


func (v *Video) addFilterVideo(filter string) *Video {
	lastFilter := v.lastVideoFilter()
	if lastFilter.FilterElement == "" {
		panic("this function is only used for adding a filter to post-existing filters")
	}
	lastFilter.addFilter(filter)
	return v
}

func (v *Video) addFilterAudio(filter string) *Video {
	lastFilter := v.lastAudioFilter()
	if lastFilter.FilterElement == "" {
		panic("this function is only used for adding a filter to post-existing filters")
	}
	lastFilter.addFilter(filter)
	return v
}

func (f *FilterComplex) addFilter(filter string) *FilterComplex {
	if f.FilterElement != "" {
		f.FilterElement += ","
	}
	f.FilterElement += filter
	return f
}

func (v *Video) lastVideoFilter() *FilterComplex {
	return &v.filterComplex[len(v.filterComplex)-1]
}

func (v *Video) lastAudioFilter() *FilterComplex {
	return &v.audio.filterComplex[len(v.audio.filterComplex)-1]
}