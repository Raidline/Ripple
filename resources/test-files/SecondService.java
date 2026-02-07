package com.bmw.trip.dips.application.service;

import java.util.List;
import com.bmw.trip.dips.infrastructure.service.InspectionZoneService;

@ApplicationScoped
public class SecondService {

    private final InspectionZoneService serv;

    public InspectionZoneService(InspectionZoneService serv) {
        this.serv = serv;
    }


    public String m1() {

        serv.updateZoneReferences(List.of());

        return "banana"
    }
}
