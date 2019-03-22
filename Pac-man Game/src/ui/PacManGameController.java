package ui;

import java.net.URL;
import java.util.ResourceBundle;

import javafx.animation.PathTransition;
import javafx.event.ActionEvent;
import javafx.fxml.FXML;
import javafx.scene.layout.Pane;
import javafx.scene.paint.Color;
import javafx.scene.shape.Arc;
import javafx.scene.shape.ArcType;
import javafx.scene.shape.Polyline;
import javafx.util.Duration;

public class PacManGameController {

    @FXML
    private ResourceBundle resources;

    @FXML
    private URL location;

    @FXML
    private Pane pane1;

    @FXML
    void initialize() {
    	pane1.setStyle("-fx-background-color: green");
    }
    
    @FXML
    void newGame2(ActionEvent event) {
    	int maxWidth = (int)pane1.getWidth();
    	int maxHeight = (int)pane1.getHeight();
    	double yRandom = 39+(Math.random()* maxHeight-64);
    	double xRandom = 39+(Math.random()* maxWidth-64);
    	
    	Arc pacman2 = new Arc(39, yRandom, 27, 26, 45, 270);
    	pacman2.setFill(Color.YELLOW);
    	pacman2.setType(ArcType.ROUND);
    	pacman2.setStrokeWidth(2); pacman2.setStroke(Color.BLACK);
    	
    	pane1.getChildren().add(pacman2);
    	
    	Polyline polyline2 = new Polyline();
    	polyline2.getPoints().addAll(new Double[] {
    			xRandom, 39.0,
    			xRandom, (double)maxHeight-39,
    			xRandom, 39.0});
    	PathTransition pathTransition = new PathTransition();
    	pathTransition.setNode(pacman2);
    	pathTransition.setDuration(Duration.seconds(5));
    	pathTransition.setPath(polyline2);
    	pathTransition.setOrientation(PathTransition.OrientationType.ORTHOGONAL_TO_TANGENT);
    	pathTransition.setCycleCount(PathTransition.INDEFINITE);
    	pathTransition.play();
    	
    }
    
    @FXML
    void newGame(ActionEvent event) {
    	int maxWidth = (int)pane1.getWidth();
    	int maxHeight = (int)pane1.getHeight();
    	double yRandom = 39+Math.random()* maxHeight-64;
    	double xRandom = 39+Math.random()* maxWidth-64;
    	
    	Arc pacman = new Arc(39, yRandom, 27, 26, 45, 270);
    	pacman.setFill(Color.YELLOW);
    	pacman.setType(ArcType.ROUND);
    	pacman.setStrokeWidth(2); pacman.setStroke(Color.BLACK);
    	
    	pane1.getChildren().add(pacman);
    	
    	Polyline polyline1 = new Polyline();
    	polyline1.getPoints().addAll(new Double[] {
    			39.0, yRandom,
    			(double)maxWidth-39, yRandom,
    			39.0, yRandom});
    	PathTransition pathTransition = new PathTransition();
    	pathTransition.setNode(pacman);
    	pathTransition.setDuration(Duration.seconds(5));
    	pathTransition.setPath(polyline1);
    	pathTransition.setOrientation(PathTransition.OrientationType.ORTHOGONAL_TO_TANGENT);
    	pathTransition.setCycleCount(PathTransition.INDEFINITE);
    	pathTransition.play();
    	
    }
    
    
}
